package test_explorer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/xgo/cmd/xgo/pathsum"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/netutil"
	"github.com/xhd2015/xgo/support/session"
)

type StartSessionRequest struct {
	Item  *TestingItem `json:"item"`
	Path  []string     `json:"path"`
	Debug bool         `json:"debug"`
	Trace bool         `json:"trace"`
}

// TODO: make FE call /session/destroy
func setupRunHandler(server *http.ServeMux, projectDir string, logConsole bool, getTestConfig func() (*TestConfig, error)) {
	sessionManager := session.NewSessionManager()

	server.HandleFunc("/session/start", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *StartSessionRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			if req == nil || req.Item == nil || req.Item.File == "" {
				return nil, netutil.ParamErrorf("requires file")
			}

			if req.Debug && req.Item.Kind != TestingItemKind_Case {
				return nil, netutil.ParamErrorf("debug not supported: %s", req.Item.Kind)
			}
			if req.Trace && req.Item.Kind != TestingItemKind_Case {
				return nil, netutil.ParamErrorf("trace not supported: %s", req.Item.Kind)
			}

			config, err := getTestConfig()
			if err != nil {
				return nil, err
			}

			id, ses, err := sessionManager.Start()
			if err != nil {
				return nil, err
			}

			absDir, err := filepath.Abs(projectDir)
			if err != nil {
				return nil, err
			}

			runSess := &runSession{
				dir:           projectDir,
				absDir:        absDir,
				goCmd:         config.GoCmd,
				env:           config.CmdEnv(),
				testFlags:     config.Flags,
				bypassGoFlags: config.BypassGoFlags,
				progArgs:      config.Args,

				pathPrefix: []string{getRootName(absDir)},

				item:  req.Item,
				path:  req.Path,
				debug: req.Debug,
				trace: req.Trace,

				logConsole: logConsole,
				session:    ses,
			}
			err = runSess.Start()
			if err != nil {
				return nil, err
			}
			return &StartSessionResult{ID: id}, nil
		})
	})

	server.HandleFunc("/session/pollStatus", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *PollSessionRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			if req.ID == "" {
				return nil, netutil.ParamErrorf("requires id")
			}
			session, err := sessionManager.Get(req.ID)
			if err != nil {
				return nil, err
			}

			events, err := session.PollEvents()
			if err != nil {
				return nil, err
			}
			// fmt.Printf("poll: %v\n", events)
			return &PollSessionResult{
				Events: convTestingEvents(events),
			}, nil
		})
	})

	server.HandleFunc("/session/destroy", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *DestroySessionRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			if req.ID == "" {
				return nil, netutil.ParamErrorf("requires id")
			}
			err = sessionManager.Destroy(req.ID)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	})
}

type testResolver struct {
	absDir     string
	dirPkgPath string
	pkgTests   sync.Map
}

func (c *testResolver) resolvePkgTestsCached(pkgPath string) ([]*TestingItem, error) {
	dirPkgPath := c.dirPkgPath
	var subPkgPath string
	if dirPkgPath != pkgPath {
		subPkgPath = getPkgSubPath(dirPkgPath, pkgPath)
		if subPkgPath == "" {
			return nil, nil
		}
	}
	v, ok := c.pkgTests.Load(subPkgPath)
	if ok {
		return v.([]*TestingItem), nil
	}
	subDir := subPkgPath
	if filepath.Separator != '/' {
		// compatible with windows
		subDir = strings.ReplaceAll(subPkgPath, string('/'), string(filepath.Separator))
	}
	results, err := resolveTests(c.absDir, filepath.Join(c.absDir, subDir))
	if err != nil {
		return nil, err
	}
	c.pkgTests.Store(subPkgPath, results)
	return results, nil
}

func (c *testResolver) resolveTestingItem(pkgPath string, name string) (*TestingItem, error) {
	testingItems, err := c.resolvePkgTestsCached(pkgPath)
	if err != nil {
		return nil, err
	}
	for _, testingItem := range testingItems {
		if testingItem.Name == name {
			return testingItem, nil
		}
	}
	return nil, nil
}

func resolveDirPkgPath(dir string) (string, error) {
	subPaths, modPath, err := goinfo.ResolveMainModule(dir)
	if err != nil {
		return "", err
	}
	if len(subPaths) == 0 {
		return modPath, nil
	}
	return modPath + "/" + strings.Join(subPaths, "/"), nil
}

func (c *runSession) Start() error {
	item := c.item
	absDir := c.absDir
	pathPrefix := c.pathPrefix
	debug := c.debug
	trace := c.trace

	begin := time.Now()
	_ = begin

	dirPkgPath, err := resolveDirPkgPath(absDir)
	if err != nil {
		return err
	}

	// record status
	pm := &pathMapping{}
	sendEvent := func(event *TestingItemEvent) {
		if event.Event == Event_ItemStatus {
			if event.Status != "" {
				pm.Set(event.Path, event.Status)
			}
		}
		event.LogConsole = c.logConsole
		c.sendEvent(event)
	}

	// notify begin
	sendEvent(&TestingItemEvent{
		Event: Event_TestStart,
	})

	paths, itemPaths, names := getTestPaths(item, pathPrefix)

	// set initial state
	for _, itemPath := range itemPaths {
		pm.Set(itemPath, RunStatus_Running)
	}

	rootPath := c.path
	if len(rootPath) == 0 {
		if len(itemPaths) > 0 {
			rootPath = itemPaths[0]
		} else {
			rootPath = pathPrefix
		}
	}

	plainMsgBuilder := func(line []byte) ([]*TestingItemEvent, error) {
		return []*TestingItemEvent{{
			Event: Event_ItemStatus,
			Path:  rootPath,
			Msg:   string(line),
		}}, nil
	}

	var singleCase bool
	var eventBuilder func(line []byte) ([]*TestingItemEvent, error)
	if item.Kind == TestingItemKind_Case {
		singleCase = true
		eventBuilder = plainMsgBuilder
	} else {
		tResolver := &testResolver{
			absDir:     absDir,
			dirPkgPath: dirPkgPath,
		}
		jsonTestEventBuilder := &jsonTestEventBuilder{
			pathPrefix:   c.pathPrefix,
			dirPkgPath:   dirPkgPath,
			testResolver: tResolver,
			pm:           pm,
		}
		eventBuilder = jsonTestEventBuilder.build
	}

	// run test
	errCh := make(chan error)
	r, w := io.Pipe()

	var defStdErrReader *io.PipeReader
	var defStdErr *io.PipeWriter
	if !singleCase {
		defStdErrReader, defStdErr = io.Pipe()
	}

	// trace
	var traceDir string

	// in go, file is ignored under a package
	// the traceDir corresponds to the test
	// case's dir
	// i.e.
	//  case = pkg/some_test.go/TestSomething/sub
	//  traceDir = ROOT/pkg
	//  caseSubPath = TestSomething/sub
	var caseSubPath string
	if singleCase && trace {
		subPath, projectRoot, err := goinfo.FindGoModDirSubPath(absDir)
		if err != nil {
			return err
		}
		itemDir := filepath.Dir(c.item.RelPath)
		if len(subPath) > 0 {
			itemDir = filepath.Join(filepath.Join(subPath...), itemDir)
		}
		traceDir, err = getConsistentTraceDir(projectRoot, itemDir)
		if err != nil {
			return err
		}
		caseSubPath = item.NameUnderPkg
		debugF("absDir=%s,itemRelPath=%s, traceDir=%s\n", absDir, c.item.RelPath, traceDir)
	}

	go func() {
		var err error
		defer func() {
			w.Close()
			errCh <- err
		}()
		var stderr io.Writer
		stdout := io.MultiWriter(os.Stdout, w)
		if singleCase {
			stderr = io.MultiWriter(os.Stderr, w)
		} else {
			stderr = io.MultiWriter(os.Stderr, defStdErr)
		}
		pathArgs := formatPathArgs(paths)
		runNames := formatRunNames(names)

		testFlags := c.testFlags
		if !debug && !singleCase {
			testFlags = append([]string{"-json"}, testFlags...)
		}
		if traceDir != "" {
			testFlags = append(testFlags, "--strace", "--strace-dir", traceDir)
		}

		if !debug {
			testArgs := joinTestArgs(pathArgs, runNames)
			err = runTest(c.goCmd, c.dir, testFlags, testArgs, c.bypassGoFlags, c.progArgs, c.env, stdout, stderr)
		} else {
			err = debugTest(c.goCmd, c.dir, item.File, testFlags, pathArgs, runNames, stdout, stderr, c.progArgs, c.env)
		}

		if err != nil {
			fmt.Printf("test err: %v\n", err)
		}
		fmt.Printf("test end\n")
	}()

	// consume stderr
	if defStdErrReader != nil {
		go func() {
			defer defStdErr.Close()
			consumeTestEvent(defStdErrReader, rootPath, plainMsgBuilder, func(event *TestingItemEvent) {
				// this loop will never write pm
				// so access to pm is single threaded
				sendEvent(event)
			})
		}()
	}

	// consume test events
	go func() {
		// wait all finish
		defer func() {
			err := <-errCh
			if err != nil {
				sendEvent(&TestingItemEvent{Event: Event_ItemStatus, Path: rootPath, Msg: err.Error(), Status: RunStatus_Fail})
			}

			// set all sub cases as success
			pm.Range(func(path []string, status RunStatus) bool {
				if status == "" || status == RunStatus_Running || status == RunStatus_NotRun {
					sendEvent(&TestingItemEvent{
						Event:  Event_ItemStatus,
						Path:   append([]string{}, path...), // copy, see https://github.com/xhd2015/xgo/issues/212
						Status: RunStatus_Success,
					})
				}
				// read trace
				if traceDir != "" {
					debugF("path: %v, baseTraceCase: %v, rootPath: %v", path, caseSubPath, rootPath)
					if suffix, ok := trimPrefix(path, rootPath); ok {
						records := readTrace(traceDir, caseSubPath, suffix)
						if records != nil {
							sendEvent(&TestingItemEvent{
								Event:        Event_UpdateTrace,
								Path:         path,
								TraceRecords: records,
							})
						}
					}
				}
				return true
			})

			sendEvent(&TestingItemEvent{
				Event: Event_TestEnd,
			})
		}()
		consumeTestEvent(r, rootPath, eventBuilder, func(event *TestingItemEvent) {
			sendEvent(event)
		})
	}()

	return nil
}

func trimPrefix(path []string, root []string) ([]string, bool) {
	if len(path) < len(root) {
		return nil, false
	}
	for i, p := range root {
		if path[i] != p {
			return nil, false
		}
	}
	return path[len(root):], true
}

func readTrace(traceDir string, baseCaseName string, subPath []string) []*CallRecord {
	traceFile := filepath.Join(traceDir, baseCaseName, filepath.Join(subPath...)) + ".json"
	debugF("traceFile: %s", traceFile)
	records, err := readTraceFromFile(traceFile)
	if err != nil {
		return []*CallRecord{
			{
				Pkg:   "github.com/xhd2015/xgo/cmd/xgo/test-explorer",
				Func:  "readTraceFromFile",
				Error: err.Error(),
			},
		}
	}
	return records
}

func readTraceFromFile(file string) (records []*CallRecord, err error) {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
	}()
	data, err := fileutil.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var rootExport *RootExport
	err = json.Unmarshal(data, &rootExport)
	if err != nil {
		return nil, err
	}
	c := &traceConverter{}
	if rootExport != nil {
		records = c.convertStacks(rootExport.Children)
	}
	return records, nil
}

const debugLog = false

func debugF(format string, args ...interface{}) {
	if !debugLog {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

type pathMapping struct {
	record *record
	m      map[string]*pathMapping
}

type record struct {
	status RunStatus
}

func (c *pathMapping) Get(path []string) (bool, RunStatus) {
	if len(path) == 0 {
		if c.record == nil {
			return false, ""
		}
		return true, c.record.status
	}
	sub := c.m[path[0]]
	if sub == nil {
		return false, ""
	}
	return sub.Get(path[1:])
}

func (c *pathMapping) Set(path []string, status RunStatus) {
	if len(path) == 0 {
		if c.record == nil {
			c.record = &record{status: status}
		} else {
			c.record.status = status
		}
		return
	}
	sub := c.m[path[0]]
	if sub == nil {
		sub = &pathMapping{
			m: make(map[string]*pathMapping),
		}
		if c.m == nil {
			c.m = make(map[string]*pathMapping, 1)
		}
		c.m[path[0]] = sub
	}
	sub.Set(path[1:], status)
}

func (c *pathMapping) Range(f func(path []string, status RunStatus) bool) {
	c.traverse(nil, f)
}

func (c *pathMapping) traverse(prefix []string, f func(path []string, status RunStatus) bool) {
	if c.record != nil {
		if !f(prefix, c.record.status) {
			return
		}
	}
	for k, v := range c.m {
		v.traverse(append(prefix, k), f)
	}
}

func runTest(goCmd string, dir string, customFlags []string, testArgs []string, bypassGoFlags bool, progArgs []string, env []string, stdout io.Writer, stderr io.Writer) error {
	if goCmd == "" {
		goCmd = "go"
	}
	testFlags := make([]string, 0, len(testArgs)+len(customFlags)+2)
	testFlags = append(testFlags, "test", "-v")
	testFlags = append(testFlags, customFlags...)
	testFlags = append(testFlags, testArgs...)
	if len(progArgs) > 0 {
		// -args after pkg can be accepted by go test
		testFlags = append(testFlags, "-args")

		// this extra '--' makes test binary skip
		// treating flags starts with '-' as
		// test flags and stop complaining
		// -something provided but not defined
		// see https://github.com/xhd2015/xgo/issues/263
		if bypassGoFlags {
			testFlags = append(testFlags, "--")
		}
		testFlags = append(testFlags, progArgs...)
	}

	return cmd.Debug().Env(env).Dir(dir).Stdout(stdout).Stderr(stderr).
		Run(goCmd, testFlags...)
}

func consumeTestEvent(r io.Reader, rootItemPath []string, builder func(line []byte) ([]*TestingItemEvent, error), callback func(*TestingItemEvent)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		data := scanner.Bytes()

		events, err := builder(data)
		if err != nil {
			callback(&TestingItemEvent{
				Event: Event_ErrorMsg,
				Path:  rootItemPath,
				Msg:   err.Error(),
			})
			return
		}
		for _, event := range events {
			callback(event)
		}
	}
}

type jsonTestEventBuilder struct {
	pathPrefix   []string
	absDir       string
	dirPkgPath   string
	testResolver *testResolver

	pm *pathMapping

	// parser
	prefix []string
}

func (c *jsonTestEventBuilder) build(line []byte) ([]*TestingItemEvent, error) {
	event, err := c.parse(line)
	if err != nil {
		return nil, err
	}
	return buildEvent(event, c.pathPrefix, c.dirPkgPath, c.pm, c.testResolver)
}

var failRegex = regexp.MustCompile(`^FAIL\s+([^\s]+)\s+.*$`)

// -json will not output json if build failed
// $ go test -json ./script/build-release
// TODO: parse std error
// stderr: # github.com/xhd2015/xgo/script/build-release [github.com/xhd2015/xgo/script/build-release.test]
// stderr: script/build-release/fixup_test.go:10:17: undefined: getGitDir
// stdout: FAIL    github.com/xhd2015/xgo/script/build-release [build failed]
func (c *jsonTestEventBuilder) parse(line []byte) (*TestEvent, error) {
	// fmt.Printf("line: %s\n", string(data))
	if bytes.HasPrefix(line, []byte{'{'}) {
		var testEvent *TestEvent
		err := json.Unmarshal(line, &testEvent)
		if err != nil {
			return nil, err
		}
		return testEvent, nil
	}
	s := string(line)
	m := failRegex.FindStringSubmatch(s)
	if m == nil {
		c.prefix = append(c.prefix, s)
		return nil, nil
	}
	pkg := m[1]
	c.prefix = nil

	output := strings.Join(c.prefix, "\n") + "\n" + s
	return &TestEvent{
		Package: pkg,
		Action:  TestEventAction_Fail,
		Output:  output,
	}, nil
}

func getConsistentTraceDir(projectDir string, subDir string) (string, error) {
	var tmpRoot string
	if runtime.GOOS != "windows" {
		// prefer /tmp on unix because it is shorter
		stat, statErr := os.Stat("/tmp")
		if statErr == nil && stat.IsDir() {
			tmpRoot = "/tmp"
		}
	}
	pathSum, err := pathsum.PathSum("", projectDir)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(tmpRoot, "xgo", "test-explorer", "trace", pathSum, subDir)
	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return "", err
	}

	return fullPath, nil
}
