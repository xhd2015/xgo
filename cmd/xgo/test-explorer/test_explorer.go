package test_explorer

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/netutil"
)

type Options struct {
	// by default go
	DefaultGoCommand string
	GoCommand        string
	ProjectDir       string
	Exclude          []string
	Flags            []string
}

func Main(args []string, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if arg == "--go-command" {
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.GoCommand = args[i+1]
			i++
			continue
		}
		if arg == "--project-dir" {
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.ProjectDir = args[i+1]
			i++
			continue
		}
		if arg == "--exclude" {
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.Exclude = append(opts.Exclude, args[i+1])
			i++
			continue
		}
		if arg == "--flag" || arg == "--flags" {
			// e.g. -parallel
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.Flags = append(opts.Flags, args[i+1])
			i++
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		return fmt.Errorf("unrecognized flag: %s", arg)
	}
	return handle(opts)
}

// NOTE: case can have sub childrens

type TestingItemKind string

const (
	TestingItemKind_Dir  = "dir"
	TestingItemKind_File = "file"
	TestingItemKind_Case = "case"
)

type RunStatus string

const (
	RunStatus_NotRun  RunStatus = "not_run"
	RunStatus_Success RunStatus = "success"
	RunStatus_Fail    RunStatus = "fail"
	RunStatus_Error   RunStatus = "error"
	RunStatus_Running RunStatus = "running"
	RunStatus_Skip    RunStatus = "skip"
)

type TestingItem struct {
	Name    string          `json:"name"`
	RelPath string          `json:"relPath"`
	File    string          `json:"file"`
	Line    int             `json:"line"`
	Kind    TestingItemKind `json:"kind"`
	Error   string          `json:"error"`

	// only if Kind==dir
	// go only
	HasTestGoFiles bool `json:"hasTestGoFiles"`

	// when filter is not
	// go only
	HasTestCases bool `json:"hasTestCases"`

	Children []*TestingItem `json:"children"`
}

type BaseRequest struct {
	Name string `json:"name"`
	File string `json:"file"`
}

type DetailRequest struct {
	*BaseRequest
	Line int `json:"line"`
}

type RunRequest struct {
	*BaseRequest
	Path    []string `json:"path"`
	Verbose bool     `json:"verbose"`
}

type RunResult struct {
	Status RunStatus `json:"status"`
	Msg    string    `json:"msg"`
}

//go:embed index.html
var indexHTML string

const apiPlaceholder = "http://localhost:8080"

func compareGoVersion(a *goinfo.GoVersion, b *goinfo.GoVersion, ignorePatch bool) int {
	if a.Major != b.Major {
		return a.Major - b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor - b.Minor
	}
	if ignorePatch {
		return 0
	}
	return a.Patch - b.Patch
}

func handle(opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}

	configFile := filepath.Join(opts.ProjectDir, "test.config.json")
	err := parseConfigAndValidate(configFile, opts)
	if err != nil {
		return err
	}

	getTestConfig := func() (*TestConfig, error) {
		conf, err := parseConfigAndMergeOptions(configFile, opts)
		if err != nil {
			return nil, fmt.Errorf("read test config:%w", err)
		}
		return conf, nil
	}

	server := &http.ServeMux{}
	var url string
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(strings.ReplaceAll(indexHTML, apiPlaceholder, url)))
	})
	server.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			q := r.URL.Query()
			dir := q.Get("dir")
			if dir == "" {
				dir = opts.ProjectDir
			}
			root, err := scanTests(dir, true, opts.Exclude)
			if err != nil {
				return nil, err
			}
			return []*TestingItem{root}, nil
		})
	})

	server.HandleFunc("/detail", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *DetailRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			if req == nil {
				req = &DetailRequest{}
			}
			if req.BaseRequest == nil {
				req.BaseRequest = &BaseRequest{}
			}
			q := r.URL.Query()
			file := q.Get("file")
			if file != "" {
				req.BaseRequest.File = file
			}
			name := q.Get("name")
			if name != "" {
				req.BaseRequest.Name = name
			}
			line := q.Get("line")
			if line != "" {
				lineNum, err := strconv.Atoi(line)
				if err != nil {
					return nil, netutil.ParamErrorf("line: %v", err)
				}
				req.Line = lineNum
			}
			return getDetail(req)
		})
	})

	server.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *RunRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			config, err := getTestConfig()
			if err != nil {
				return nil, err
			}

			return run(req, config.GoCmd, config.CmdEnv(), config.Flags)
		})
	})

	setupSessionHandler(server, opts.ProjectDir, getTestConfig)

	server.HandleFunc("/openVscode", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			q := r.URL.Query()
			file := q.Get("file")
			line := q.Get("line")

			err := cmd.Run("code", "--goto", fmt.Sprintf("%s:%s", file, line))
			return nil, err
		})
	})

	return netutil.ServePortHTTP(server, 7070, true, 500*time.Millisecond, func(port int) {
		url = fmt.Sprintf("http://localhost:%d", port)
		fmt.Printf("Server listen at %s\n", url)
		openURL(url)
	})
}

func openURL(url string) {
	openCmd := "open"
	if runtime.GOOS == "windows" {
		openCmd = "explorer"
	}
	cmd.Run(openCmd, url)
}

func parseBody(r io.Reader, req interface{}) error {
	if r == nil {
		return nil
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, req)
}

func scanTests(dir string, needParseTests bool, exclude []string) (*TestingItem, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	root := &TestingItem{
		Name: filepath.Base(absDir),
		File: absDir,
		Kind: TestingItemKind_Dir,
	}
	itemMapping := make(map[string]*TestingItem)
	itemMapping[absDir] = root

	getParent := func(path string) (*TestingItem, error) {
		parent := itemMapping[filepath.Dir(path)]
		if parent == nil {
			return nil, fmt.Errorf("item mapping not found: %s", filepath.Dir(path))
		}
		return parent, nil
	}
	err = fileutil.WalkRelative(absDir, func(path, relPath string, d fs.DirEntry) error {
		if relPath == "" {
			return nil
		}
		if len(exclude) > 0 {
			var found bool
			for _, e := range exclude {
				if e == relPath {
					found = true
					break
				}
			}
			if found {
				if d.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
		}
		if d.IsDir() {
			// vendor inside root
			if relPath == "vendor" {
				return filepath.SkipDir
			}

			hasGoMod, err := fileutil.FileExists(filepath.Join(path, "go.mod"))
			if err != nil {
				return err
			}
			if hasGoMod {
				// sub project
				return filepath.SkipDir
			}
			parent, err := getParent(path)
			if err != nil {
				return err
			}
			item := &TestingItem{
				Name:    filepath.Base(relPath),
				RelPath: relPath,
				File:    path,
				Kind:    TestingItemKind_Dir,
			}
			itemMapping[path] = item
			parent.Children = append(parent.Children, item)
			return nil
		}

		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		parent, err := getParent(path)
		if err != nil {
			return err
		}
		item := &TestingItem{
			Name:    filepath.Base(relPath),
			RelPath: relPath,
			File:    path,
			Kind:    TestingItemKind_File,
		}
		itemMapping[path] = item
		parent.HasTestGoFiles = true
		parent.Children = append(parent.Children, item)

		if needParseTests {
			tests, parseErr := parseTests(path)
			if parseErr != nil {
				item.Error = parseErr.Error()
			} else {
				for _, test := range tests {
					test.RelPath = relPath
				}
				// TODO: what if test case name same with sub dir?
				item.Children = append(item.Children, tests...)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// filter items without
	// any tests
	filterItem(root, needParseTests)
	return root, nil
}

type DetailResponse struct {
	Content string `json:"content"`
}

func getDetail(req *DetailRequest) (*DetailResponse, error) {
	if req == nil || req.BaseRequest == nil || req.File == "" {
		return nil, netutil.ParamErrorf("requires file")
	}
	if req.Name == "" {
		return nil, netutil.ParamErrorf("requires name")
	}

	fset, decls, err := parseTestFuncs(req.File)
	if err != nil {
		return nil, err
	}
	var found *ast.FuncDecl
	for _, decl := range decls {
		if decl.Name != nil && decl.Name.Name == req.Name {
			found = decl
			break
		}
	}
	if found == nil {
		return nil, netutil.ParamErrorf("not found: %s", req.Name)
	}
	content, err := ioutil.ReadFile(req.File)
	if err != nil {
		return nil, err
	}
	i := fset.Position(found.Pos()).Offset
	j := fset.Position(found.End()).Offset
	return &DetailResponse{
		Content: string(content)[i:j],
	}, nil
}
func run(req *RunRequest, goCmd string, env []string, testFlags []string) (*RunResult, error) {
	if req == nil || req.BaseRequest == nil || req.File == "" {
		return nil, fmt.Errorf("requires file")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("requires name")
	}
	// fmt.Printf("run:%v\n", req)
	var buf bytes.Buffer
	args := []string{"test", "-run", fmt.Sprintf("^%s$", req.Name)}
	if req.Verbose {
		args = append(args, "-v")
	}
	args = append(args, testFlags...)
	if goCmd == "" {
		goCmd = "go"
	}

	fmt.Printf("test: %s %v\n", goCmd, args)
	runErr := cmd.Dir(filepath.Dir(req.File)).
		Env(env).
		Stderr(io.MultiWriter(os.Stderr, &buf)).
		Stdout(io.MultiWriter(os.Stdout, &buf)).
		Run(goCmd, args...)
	if runErr != nil {
		return &RunResult{
			Status: RunStatus_Fail,
			Msg:    buf.String(),
		}, nil
	}

	return &RunResult{
		Status: RunStatus_Success,
		Msg:    buf.String(),
	}, nil
}

func filterItem(item *TestingItem, withCases bool) *TestingItem {
	if item == nil {
		return nil
	}

	if withCases {
		children := item.Children
		n := len(children)
		i := 0
		for j := 0; j < n; j++ {
			child := filterItem(children[j], withCases)
			if child != nil {
				children[i] = child
				i++
			}
		}
		item.Children = children[:i]
		if i == 0 && item.Kind != TestingItemKind_Case {
			return nil
		}
	} else {
		if !item.HasTestCases {
			return nil
		}
	}

	return item
}

func parseTests(file string) ([]*TestingItem, error) {
	fset, decls, err := parseTestFuncs(file)
	if err != nil {
		return nil, err
	}
	items := make([]*TestingItem, 0, len(decls))
	for _, fnDecl := range decls {
		items = append(items, &TestingItem{
			Name: fnDecl.Name.Name,
			File: file,
			Line: fset.Position(fnDecl.Pos()).Line,
			Kind: TestingItemKind_Case,
		})
	}
	return items, nil
}

func parseTestFuncs(file string) (*token.FileSet, []*ast.FuncDecl, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	var results []*ast.FuncDecl
	for _, decl := range astFile.Decls {
		fnDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fnDecl.Name == nil {
			continue
		}
		if !strings.HasPrefix(fnDecl.Name.Name, "Test") {
			continue
		}
		if fnDecl.Body == nil {
			continue
		}
		if fnDecl.Type.Params == nil || len(fnDecl.Type.Params.List) != 1 {
			continue
		}
		results = append(results, fnDecl)
	}
	return fset, results, nil
}
