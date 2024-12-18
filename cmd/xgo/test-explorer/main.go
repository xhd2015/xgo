package test_explorer

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/ast"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/xgo/cmd/xgo/coverage/cov_control"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/git"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/load/loadcov"
	"github.com/xhd2015/xgo/cmd/xgo/test-explorer/icov"
	"github.com/xhd2015/xgo/support/flag"
	"github.com/xhd2015/xgo/support/pattern"

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
	Args             []string

	// can be true,false
	Coverage string
	// --coverage-profile string
	CoverageProfile  string
	CoverageDiffWith string

	Config string
	Port   string
	Bind   string

	LogConsole bool

	Verbose bool
}

func Main(args []string, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	var flagHelp bool
	n := len(args)
	var remainArgs []string
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if arg == "-h" || arg == "--help" {
			flagHelp = true
			continue
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
		if arg == "--arg" || arg == "--args" {
			// e.g. -parallel
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.Args = append(opts.Args, args[i+1])
			i++
			continue
		}
		if arg == "--coverage" {
			opts.Coverage = "true"
			continue
		}
		if strings.HasPrefix(arg, "--coverage=") {
			v := strings.TrimPrefix(arg, "--coverage=")
			if v == "true" {
				opts.Coverage = "true"
			} else if v == "false" {
				opts.Coverage = "false"
			} else {
				return fmt.Errorf("invalid: %s", arg)
			}
			continue
		}
		if arg == "--coverage-profile" {
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.CoverageProfile = args[i+1]
			i++
			continue
		}
		if arg == "--coverage-diff-with" {
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.CoverageDiffWith = args[i+1]
			i++
			continue
		}
		if arg == "-v" || arg == "--verbose" {
			opts.Verbose = true
			continue
		}
		if arg == "--log-console" {
			opts.LogConsole = true
			continue
		}

		ok, err := flag.TryParseFlagValue("--config", &opts.Config, nil, &i, args)
		if err != nil {
			return err
		}
		if ok {
			continue
		}
		if arg == "--port" {
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.Port = args[i+1]
			i++
			continue
		}
		if arg == "--bind" {
			if i+1 >= n {
				return fmt.Errorf("%s requires value", arg)
			}
			opts.Bind = args[i+1]
			i++
			continue
		}

		if !strings.HasPrefix(arg, "-") {
			remainArgs = append(remainArgs, arg)
			continue
		}
		return fmt.Errorf("unrecognized flag: %s", arg)
	}
	if flagHelp || (len(remainArgs) > 0 && remainArgs[0] == "help") {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}
	return handle(opts, remainArgs)
}

// NOTE: case can have sub childrens

type TestingItemKind string

const (
	TestingItemKind_Dir  = "dir"
	TestingItemKind_File = "file"
	TestingItemKind_Case = "case"
)

func (c TestingItemKind) Order() int {
	switch c {
	case TestingItemKind_Dir:
		return 0
	case TestingItemKind_File:
		return 1
	case TestingItemKind_Case:
		return 2
	default:
		return -1
	}
}

type RunStatus string

const (
	RunStatus_NotRun  RunStatus = "not_run"
	RunStatus_Success RunStatus = "success"
	RunStatus_Fail    RunStatus = "fail"
	RunStatus_Error   RunStatus = "error"
	RunStatus_Running RunStatus = "running"
	RunStatus_Skip    RunStatus = "skip"
)

// TestingItem describe a test item
// when invoked from sub dir, relPath
// is relative to the sub dir
type TestingItem struct {
	Key string `json:"key"`
	// base name, e.g.: subcase if full name is TestSomething/subcase
	Name string `json:"name"`
	// base case name, e.g.:  TestSomething
	BaseCaseName string `json:"baseCaseName"` // the base case's name
	// full name, e.g.: TestSomething/subcase
	NameUnderPkg string          `json:"nameUnderPkg"` // the name under pkg
	RelPath      string          `json:"relPath"`
	File         string          `json:"file"`
	Line         int             `json:"line"`
	Kind         TestingItemKind `json:"kind"`
	Error        string          `json:"error"`

	// only if Kind==dir
	// indicating any file ends with _test.go
	// go only
	HasTestGoFiles bool `json:"hasTestGoFiles"`

	// valid for Kind==dir,file
	// indicating any cases belongs to this item
	// go only
	HasTestCases bool              `json:"hasTestCases"`
	State        *TestingItemState `json:"state"`

	Children []*TestingItem `json:"children"`
}

// clone excluding children
func (c *TestingItem) CloneSelf() *TestingItem {
	if c == nil {
		return nil
	}
	return &TestingItem{
		Key:            c.Key,
		Name:           c.Name,
		BaseCaseName:   c.BaseCaseName,
		NameUnderPkg:   c.NameUnderPkg,
		RelPath:        c.RelPath,
		File:           c.File,
		Line:           c.Line,
		Kind:           c.Kind,
		Error:          c.Error,
		HasTestGoFiles: c.HasTestGoFiles,
		HasTestCases:   c.HasTestCases,
		State:          c.State.Clone(),
	}
}

type HideType string

const (
	HideType_None     HideType = ""
	HideType_All      HideType = "all"
	HideType_Children HideType = "children"
)

type TestingItemState struct {
	Selected  bool      `json:"selected"`
	Expanded  bool      `json:"expanded"`
	Status    RunStatus `json:"status"`
	Debugging bool      `json:"debugging"`
	Logs      string    `json:"logs"`
	HideType  HideType  `json:"hideType"`
}

func (c *TestingItemState) Clone() *TestingItemState {
	if c == nil {
		return nil
	}
	return &TestingItemState{
		Selected:  c.Selected,
		Expanded:  c.Expanded,
		Status:    c.Status,
		Debugging: c.Debugging,
		Logs:      c.Logs,
		HideType:  c.HideType,
	}
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

func handle(opts *Options, args []string) error {
	if opts == nil {
		opts = &Options{}
	}
	projectDir := opts.ProjectDir
	subPath, projectRoot, err := goinfo.FindGoModDirSubPath(projectDir)
	if err != nil {
		return err
	}
	var configFile string
	configFileName := opts.Config
	var configFileRequired bool
	if configFileName != "none" {
		if configFileName == "" {
			configFile = filepath.Join(projectRoot, "test.config.json")
		} else {
			configFileRequired = true
			if filepath.IsAbs(configFileName) {
				configFile = configFileName
			} else {
				configFile = filepath.Join(projectRoot, configFileName)
			}
		}
		err := parseConfigAndValidate(configFile, opts, configFileRequired)
		if err != nil {
			return err
		}
	}

	getTestConfig := func() (*TestConfig, error) {
		conf, err := parseConfigAndMergeOptions(configFile, opts, configFileRequired)
		if err != nil {
			return nil, fmt.Errorf("read test config:%w", err)
		}
		return conf, nil
	}
	// headless mode
	conf, err := getTestConfig()
	if err != nil {
		return err
	}
	// create coverage controller
	var covController icov.Controller
	if conf.Coverage == nil || !conf.Coverage.Disabled {
		var diffWith string
		var profile string
		if conf.Coverage != nil {
			profile = conf.Coverage.Profile
			diffWith = conf.Coverage.DiffWith
		}
		if diffWith == "" {
			diffWith = "origin/master"
		}
		covController, err = cov_control.New(projectDir, nil, diffWith, profile)
		if err != nil {
			return fmt.Errorf("init coverage: %w", err)
		}
	}
	if conf.Xgo != nil && conf.Xgo.AutoUpdate && os.Getenv("XGO_AUTO_UPDATE") != "never" {
		autoUpdateXgo()
	}
	var coverageFlags []string
	if covController != nil {
		// example: -cover -coverpkg xxx -coverprofile coverage.out
		modPath, err := goinfo.GetModPath(projectDir)
		if err != nil {
			return err
		}
		coverageFlags = append(coverageFlags, "-cover", "-coverpkg", modPath+"/...")
		profilePath, err := covController.GetCoverageProfilePath()
		if err != nil {
			return err
		}
		if profilePath != "" {
			// -cover -coverpkg
			coverageFlags = append(coverageFlags, "-coverprofile", profilePath)
		}
	}
	if len(args) > 0 && args[0] == "test" {
		// headless mod
		remainArgs := args[1:]
		var testArgs []string
		if len(remainArgs) > 0 {
			testArgs = remainArgs
		} else {
			root, err := scanTests(projectRoot, subPath, true, conf.Exclude)
			if err != nil {
				return err
			}
			paths, _, names := getTestPaths(root, nil)
			pathArgs := formatPathArgs(paths)
			runNames := formatRunNames(names)
			testArgs = joinTestArgs(pathArgs, runNames)
		}

		return runTest(conf.GoCmd, projectDir, appendCopy(conf.Flags, coverageFlags...), testArgs, conf.BypassGoFlags, conf.Args, conf.CmdEnv(), nil, nil)
	}

	server := &http.ServeMux{}
	var url string
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		if uri != "" && uri != "/" {
			w.WriteHeader(404)
			w.Write([]byte("requested source not found:" + uri))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(strings.ReplaceAll(indexHTML, apiPlaceholder, url)))
	})
	server.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			conf, err := getTestConfig()
			if err != nil {
				return nil, err
			}
			root, err := scanTests(projectRoot, subPath, true, conf.Exclude)
			if err != nil {
				return nil, err
			}
			return root, nil
		})
	})

	var actualPort int

	var covOpts loadcov.LoadAllOptions
	if covController != nil {
		profilePath, err := covController.GetCoverageProfilePath()
		if err != nil {
			return err
		}
		var diffWith string
		var include []string
		var exclude []string
		if conf.Coverage != nil {
			diffWith = conf.Coverage.DiffWith
			include = conf.Coverage.Include
			exclude = conf.Coverage.Exclude
		}
		if diffWith == "" {
			diffWith = "origin/master"
		}
		covOpts = loadcov.LoadAllOptions{
			Dir:      projectDir,
			Args:     nil,
			Profiles: []string{profilePath},
			Ref:      git.COMMIT_WORKING,
			DiffBase: diffWith,
			Include:  include,
			Exclude:  exclude,
		}
	}

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

	setupSessionHandler(server, projectDir, opts.LogConsole, getTestConfig, covController, coverageFlags)
	setupCoverageHandler(server, covController, covOpts, func() int {
		return actualPort
	})
	setupOpenHandler(server)

	// deprecated
	setupDebugHandler(server, projectDir, getTestConfig)
	setupRunHandler(server, projectDir, getTestConfig)
	host, port := netutil.GetHostAndIP(opts.Bind, opts.Port)

	autoIncrPort := true
	return netutil.ServePortHTTP(server, host, port, autoIncrPort, 500*time.Millisecond, func(port int) {
		actualPort = port
		url, extra := netutil.GetURLToOpen(host, port)
		netutil.PrintUrls(url, extra...)
		openURL(url)
	})
}

func autoUpdateXgo() {
	// auto update
	done := make(chan struct{})
	go func() {
		cmd.Debug().Run("go", "install", "github.com/xhd2015/xgo/cmd/xgo@latest")
		close(done)
	}()
	// at most wait 2 second
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
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

func getRootName(absDir string) string {
	return filepath.Base(absDir)
}

// needParseTests set to true when calling /list
func scanTests(projectRoot string, subPath []string, needParseTests bool, exclude []string) (*TestingItem, error) {
	absDir, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, err
	}

	var subPathJoin string
	if len(subPath) > 0 {
		subPathJoin = filepath.Join(subPath...)
	}
	walkDir := absDir
	if subPathJoin != "" {
		walkDir = filepath.Join(absDir, subPathJoin)
	}

	name := getRootName(walkDir)
	root := &TestingItem{
		Key:  name,
		Name: name,
		File: walkDir,
		Kind: TestingItemKind_Dir,
	}
	itemMapping := make(map[string]*TestingItem)
	itemMapping[walkDir] = root

	getParent := func(path string) (*TestingItem, error) {
		parent := itemMapping[filepath.Dir(path)]
		if parent == nil {
			return nil, fmt.Errorf("item mapping not found: %s", filepath.Dir(path))
		}
		return parent, nil
	}

	excludePatterns := pattern.CompilePatterns(exclude)
	err = fileutil.WalkRelative(walkDir, func(path, relPath string, d fs.DirEntry) error {
		rootRelPath := relPath
		if subPathJoin != "" {
			rootRelPath = filepath.Join(subPathJoin, relPath)
		}
		if rootRelPath == "" {
			return nil
		}
		if len(exclude) > 0 {
			matchPath := rootRelPath
			if filepath.Separator != '/' {
				matchPath = strings.ReplaceAll(rootRelPath, string(filepath.Separator), "/")
			}
			if excludePatterns.MatchAnyPrefix(matchPath) {
				if d.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
		}
		if relPath == "" {
			return nil
		}
		if d.IsDir() {
			// vendor inside root
			if rootRelPath == "vendor" {
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
			name := filepath.Base(rootRelPath)
			item := &TestingItem{
				Key:     name,
				Name:    name,
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
		name := filepath.Base(relPath)
		item := &TestingItem{
			Key:     name,
			Name:    name,
			RelPath: relPath,
			File:    path,
			Kind:    TestingItemKind_File,
		}
		itemMapping[path] = item
		parent.HasTestGoFiles = true
		parent.Children = append(parent.Children, item)

		if needParseTests {
			tests, parseErr := parseTests(absDir, path)
			if parseErr != nil {
				item.Error = parseErr.Error()
			} else {
				for _, test := range tests {
					test.RelPath = relPath
				}
				// TODO: what if test case name same with sub dir?
				item.Children = append(item.Children, tests...)

				if len(tests) > 0 {
					item.HasTestCases = true
					parent.HasTestCases = true
				}
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
	sortByKindAndName(root)
	selectFirstOne(root)
	return root, nil
}

func sortByKindAndName(item *TestingItem) {
	if item == nil {
		return
	}
	sort.Slice(item.Children, func(i, j int) bool {
		a := item.Children[i]
		b := item.Children[j]
		if a.Kind != b.Kind {
			return a.Kind.Order() < b.Kind.Order()
		}
		if a.Kind == TestingItemKind_Case {
			// case does sort by index
			return i < j
		}
		return strings.Compare(a.Name, b.Name) < 0
	})
	for _, child := range item.Children {
		sortByKindAndName(child)
	}
}

func selectFirstOne(item *TestingItem) bool {
	if item == nil {
		return false
	}

	if item.Kind == TestingItemKind_Case {
		if item.State == nil {
			item.State = &TestingItemState{}
		}
		item.State.Selected = true
		return true
	}

	for _, child := range item.Children {
		if selectFirstOne(child) {
			return true
		}
	}
	return false
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
	content, err := fileutil.ReadFile(req.File)
	if err != nil {
		return nil, err
	}
	i := fset.Position(found.Pos()).Offset
	j := fset.Position(found.End()).Offset
	return &DetailResponse{
		Content: string(content)[i:j],
	}, nil
}
