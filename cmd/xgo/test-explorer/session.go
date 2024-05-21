package test_explorer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/netutil"
)

type StartSessionRequest struct {
	*TestingItem
}
type StartSessionResult struct {
	ID string `json:"id"`
}
type Event string

const (
	Event_ItemStatus Event = "item_status"
	Event_Output     Event = "output"
	Event_ErrorMsg   Event = "error_msg"
	Event_TestStart  Event = "test_start"
	Event_TestEnd    Event = "test_end"
)

type TestingItemEvent struct {
	Event  Event        `json:"event"`
	Item   *TestingItem `json:"item"`
	Status RunStatus    `json:"status"`
	Msg    string       `json:"msg"`
}

type PollSessionRequest struct {
	ID string `json:"id"`
}

type PollSessionResult struct {
	Events []*TestingItemEvent `json:"events"`
}

type session struct {
	dir       string
	goCmd     string
	exclude   []string
	env       []string
	testFlags []string

	item *TestingItem

	eventCh chan *TestingItemEvent
}

func getRelDirs(root *TestingItem, file string) []string {
	var find func(t *TestingItem) *TestingItem
	find = func(t *TestingItem) *TestingItem {
		if t.File == file {
			return t
		}
		for _, child := range t.Children {
			e := find(child)
			if e != nil {
				return e
			}
		}
		return nil
	}
	target := find(root)
	if target == nil {
		return nil
	}

	var getRelPaths func(t *TestingItem) []string
	getRelPaths = func(t *TestingItem) []string {
		var dirs []string
		if t.Kind == TestingItemKind_Dir && t.HasTestGoFiles {
			dirs = append(dirs, t.RelPath)
		}
		for _, e := range t.Children {
			dirs = append(dirs, getRelPaths(e)...)
		}
		return dirs
	}
	return getRelPaths(target)
}

// see https://pkg.go.dev/cmd/test2json#hdr-Output_Format
type TestEventAction string

const (
	TestEventAction_Start  TestEventAction = "start"
	TestEventAction_Run    TestEventAction = "run"
	TestEventAction_Pass   TestEventAction = "pass"
	TestEventAction_Pause  TestEventAction = "pause"
	TestEventAction_Cont   TestEventAction = "cont"
	TestEventAction_Bench  TestEventAction = "bench"
	TestEventAction_Output TestEventAction = "output"
	TestEventAction_Fail   TestEventAction = "fail"
	TestEventAction_Skip   TestEventAction = "skip"
)

// from go/cmd/test2json
type TestEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  TestEventAction
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

func getPkgSubDirPath(modPath string, pkgPath string) string {
	// NOTE: pkgPath can be command-line-arguments
	if !strings.HasPrefix(pkgPath, modPath) {
		return ""
	}
	return strings.TrimPrefix(pkgPath[len(modPath):], "/")
}

func resolveTests(fullSubDir string) ([]*TestingItem, error) {
	files, err := os.ReadDir(fullSubDir)
	if err != nil {
		return nil, err
	}
	var results []*TestingItem
	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		if file.IsDir() {
			continue
		}
		fullFile := filepath.Join(fullSubDir, fileName)
		tests, err := parseTests(fullFile)
		if err != nil {
			return nil, err
		}
		results = append(results, tests...)
	}
	return results, nil
}

func (c *session) Start() error {
	absDir, err := filepath.Abs(c.dir)
	if err != nil {
		return err
	}
	// find all tests
	modPath, err := goinfo.GetModPath(absDir)
	if err != nil {
		return err
	}

	finish := func() {
		c.sendEvent(&TestingItemEvent{
			Event: Event_TestEnd,
		})
	}

	var testArgs []string
	file := c.item.File

	isFile, err := fileutil.IsFile(file)
	if err != nil {
		return err
	}
	if isFile {
		relPath, err := filepath.Rel(absDir, file)
		if err != nil {
			return err
		}
		var subCaseNames []string
		if c.item.Kind != TestingItemKind_Case {
			subCases, err := parseTests(file)
			if err != nil {
				return err
			}
			if len(subCases) == 0 {
				finish()
				return nil
			}
			subCaseNames = make([]string, 0, len(subCases))
			for _, subCase := range subCases {
				subCaseNames = append(subCaseNames, subCase.Name)
			}
		} else {
			subCaseNames = append(subCaseNames, c.item.Name)
		}
		// fmt.Printf("sub cases: %v\n", subCaseNames)
		testArgs = append(testArgs, "-run", fmt.Sprintf("^%s$", strings.Join(subCaseNames, "|")))
		testArgs = append(testArgs, "./"+filepath.Dir(relPath))
	} else {
		// all sub dirs
		root, err := scanTests(absDir, false, c.exclude)
		if err != nil {
			return err
		}

		// find all relDirs
		relDirs := getRelDirs(root, file)
		if len(relDirs) == 0 {
			return nil
		}
		// must exclude non packages
		// no Go files in /Users/xhd2015/Projects/xhd2015/xgo-test-explorer/support
		// fmt.Printf("dirs: %v\n", relDirs)
		for _, relDir := range relDirs {
			testArgs = append(testArgs, "./"+relDir)
		}
	}

	var pkgTests sync.Map

	resolvePkgTestsCached := func(absDir string, modPath string, pkgPath string) ([]*TestingItem, error) {
		subDir := getPkgSubDirPath(modPath, pkgPath)
		if subDir == "" {
			return nil, nil
		}
		v, ok := pkgTests.Load(subDir)
		if ok {
			return v.([]*TestingItem), nil
		}
		results, err := resolveTests(filepath.Join(absDir, subDir))
		if err != nil {
			return nil, err
		}
		pkgTests.Store(subDir, results)
		return results, nil
	}

	resolveTestFile := func(absDir, pkgPath string, name string) (string, error) {
		testingItems, err := resolvePkgTestsCached(absDir, modPath, pkgPath)
		if err != nil {
			return "", err
		}
		for _, testingItem := range testingItems {
			if testingItem.Name == name {
				return testingItem.File, nil
			}
		}
		return "", nil
	}

	c.sendEvent(&TestingItemEvent{
		Event: Event_TestStart,
	})

	r, w := io.Pipe()
	go func() {
		defer finish()
		goCmd := "go"
		if c.goCmd != "" {
			goCmd = c.goCmd
		}
		testFlags := append([]string{"test", "-json"}, c.testFlags...)
		testFlags = append(testFlags, testArgs...)
		fmt.Printf("%s %v\n", goCmd, testFlags)

		err := cmd.Env(c.env).Dir(c.dir).
			Stdout(io.MultiWriter(os.Stdout, w)).
			Run(goCmd, testFlags...)
		if err != nil {
			fmt.Printf("test err: %v\n", err)
			c.sendEvent(&TestingItemEvent{Event: Event_ErrorMsg, Msg: err.Error()})
		}
		fmt.Printf("test end\n")
	}()

	// -json will not output json if build failed
	// $ go test -json ./script/build-release
	// TODO: parse std error
	// stderr: # github.com/xhd2015/xgo/script/build-release [github.com/xhd2015/xgo/script/build-release.test]
	// stderr: script/build-release/fixup_test.go:10:17: undefined: getGitDir
	// stdout: FAIL    github.com/xhd2015/xgo/script/build-release [build failed]
	reg := regexp.MustCompile(`^FAIL\s+([^\s]+)\s+.*$`)
	go func() {
		scanner := bufio.NewScanner(r)

		var prefix []string
		for scanner.Scan() {
			var testEvent TestEvent
			data := scanner.Bytes()
			// fmt.Printf("line: %s\n", string(data))
			if !bytes.HasPrefix(data, []byte{'{'}) {
				s := string(data)
				m := reg.FindStringSubmatch(s)
				if m == nil {
					prefix = append(prefix, s)
					continue
				}
				pkg := m[1]
				prefix = nil

				output := strings.Join(prefix, "\n") + "\n" + s
				testEvent = TestEvent{
					Package: pkg,
					Action:  TestEventAction_Fail,
					Output:  output,
				}
			} else {
				err := json.Unmarshal(data, &testEvent)
				if err != nil {
					// emit global message
					fmt.Printf("err:%s %v\n", data, err)
					c.sendEvent(&TestingItemEvent{Event: Event_ErrorMsg, Msg: err.Error()})
					continue
				}
			}
			itemEvent := buildEvent(&testEvent, absDir, modPath, resolveTestFile, getPkgSubDirPath)
			if itemEvent != nil {
				c.sendEvent(itemEvent)
			}
		}
	}()

	return nil
}

func buildEvent(testEvent *TestEvent, absDir string, modPath string, resolveTestFile func(absDir string, pkgPath string, name string) (string, error), getPkgSubDirPath func(modPath string, pkgPath string) string) *TestingItemEvent {
	var kind TestingItemKind
	var fullFile string
	var status RunStatus

	if testEvent.Package != "" {
		if testEvent.Test != "" {
			kind = TestingItemKind_Case
			fullFile, _ = resolveTestFile(absDir, testEvent.Package, testEvent.Test)
		} else {
			kind = TestingItemKind_Dir
			subDir := getPkgSubDirPath(modPath, testEvent.Package)
			if subDir != "" {
				fullFile = filepath.Join(absDir, subDir)
			}
		}
	}

	switch testEvent.Action {
	case TestEventAction_Run:
		status = RunStatus_Running
	case TestEventAction_Pass:
		status = RunStatus_Success
	case TestEventAction_Fail:
		status = RunStatus_Fail
	case TestEventAction_Skip:
		status = RunStatus_Skip
	}
	return &TestingItemEvent{
		Event: Event_ItemStatus,
		Item: &TestingItem{
			Kind: kind,
			File: fullFile,
			Name: testEvent.Test,
		},
		Status: status,
		Msg:    testEvent.Output,
	}
}

func (c *session) Poll() []*TestingItemEvent {
	var events []*TestingItemEvent

	timeout := time.After(5 * time.Second)
	for {
		select {
		case event := <-c.eventCh:
			events = append(events, event)
		case <-timeout:
			return events
		default:
			if len(events) > 0 {
				return events
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *session) sendEvent(event *TestingItemEvent) {
	c.eventCh <- event
}

// TODO: add /session/destroy
func setupSessionHandler(server *http.ServeMux, projectDir string, getTestConfig func() (*TestConfig, error)) {
	var nextID int64 = 0
	var sessionMapping sync.Map

	server.HandleFunc("/session/start", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			var req *StartSessionRequest
			err := parseBody(r.Body, &req)
			if err != nil {
				return nil, err
			}
			if req == nil || req.TestingItem == nil || req.File == "" {
				return nil, netutil.ParamErrorf("requires file")
			}

			config, err := getTestConfig()
			if err != nil {
				return nil, err
			}

			idInt := atomic.AddInt64(&nextID, 1)
			// to avoid stale requests from older pages
			id := fmt.Sprintf("session_%s_%d", time.Now().Format("2006-01-02_15:04:05"), idInt)

			sess := &session{
				dir:       projectDir,
				goCmd:     config.GoCmd,
				exclude:   config.Exclude,
				env:       config.CmdEnv(),
				testFlags: config.Flags,

				eventCh: make(chan *TestingItemEvent, 100),
				item:    req.TestingItem,
			}
			sessionMapping.Store(id, sess)
			err = sess.Start()
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
			val, ok := sessionMapping.Load(req.ID)
			if !ok {
				return nil, netutil.ParamErrorf("session %s does not exist or has been removed", req.ID)
			}
			sess := val.(*session)

			events := sess.Poll()
			// fmt.Printf("poll: %v\n", events)
			return &PollSessionResult{
				Events: events,
			}, nil
		})
	})
}
