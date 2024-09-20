package test_explorer

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/session"
)

func setupTestHandler(server *http.ServeMux, projectDir string, getTestConfig func() (*TestConfig, error)) {
	setupPollHandler(server, "/run", projectDir, getTestConfig, run)
}

type StartSessionResult struct {
	ID string `json:"id"`
}

type Event string

const (
	Event_ItemStatus  Event = "item_status"
	Event_MergeTree   Event = "merge_tree"
	Event_Output      Event = "output"
	Event_ErrorMsg    Event = "error_msg"
	Event_TestStart   Event = "test_start"
	Event_TestEnd     Event = "test_end"
	Event_UpdateTrace Event = "update_trace"
)

type TestingItemEvent struct {
	Event        Event         `json:"event"`
	Item         *TestingItem  `json:"item"`
	Path         []string      `json:"path"`
	Status       RunStatus     `json:"status"`
	Msg          string        `json:"msg"`
	LogConsole   bool          `json:"logConsole"`
	TraceRecords []*CallRecord `json:"traceRecords"`
}

type PollSessionRequest struct {
	ID string `json:"id"`
}

type PollSessionResult struct {
	Events []*TestingItemEvent `json:"events"`
}
type DestroySessionRequest struct {
	ID string `json:"id"`
}

type runSession struct {
	dir       string
	absDir    string
	goCmd     string
	env       []string
	testFlags []string

	bypassGoFlags bool
	progArgs      []string

	pathPrefix []string

	item  *TestingItem
	path  []string
	debug bool
	trace bool // xgo stack trace

	logConsole bool

	session session.Session
}

func formatPathArgs(paths []string) []string {
	args := make([]string, 0, len(paths))
	for _, relPath := range paths {
		if filepath.Separator != '/' {
			relPath = strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		}
		args = append(args, "./"+relPath)
	}
	return args
}
func formatRunNames(names []string) string {
	if names == nil {
		return ""
	}
	return fmt.Sprintf("^%s$", strings.Join(escapeRegexNames(names), "|"))
}

func escapeRegexNames(names []string) []string {
	replacedNames := make([]string, 0, len(names))
	for _, name := range names {
		replacedNames = append(replacedNames, escapeRegexName(name))
	}
	return replacedNames
}

var replacer = strings.NewReplacer(
	// ".", "\\.",
	"{", "\\{",
	"}", "\\}",
	"[", "\\[",
	"]", "\\]",
	"|", "\\|",
)

func escapeRegexName(name string) string {
	return replacer.Replace(name)
}

func joinTestArgs(pathArgs []string, runNames string) []string {
	args := make([]string, 0, len(pathArgs))
	if runNames != "" {
		args = append(args, "-run", runNames)
	}
	args = append(args, pathArgs...)
	return args
}

func getTestPaths(item *TestingItem, pathPrefix []string) (paths []string, itemPaths [][]string, names []string) {
	switch item.Kind {
	case TestingItemKind_Case:
		testName := item.NameUnderPkg
		if testName == "" {
			testName = item.Name
		}
		relPath := filepath.Dir(item.RelPath)
		paths = []string{relPath}
		names = []string{testName}

		fileItemPath := getCaseItemPath(pathPrefix, item.RelPath, item.Name, item.NameUnderPkg)
		itemPaths = [][]string{fileItemPath}
	case TestingItemKind_File:
		relPath, cases := getFileSubCases(item)
		paths = []string{relPath}

		names = cases
		if names == nil {
			names = make([]string, 0)
		}

		fileItemPath := getFileItemPath(pathPrefix, item.RelPath)
		itemPaths = append(itemPaths, fileItemPath)
		for _, c := range cases {
			itemPaths = append(itemPaths, appendCopy(fileItemPath, c))
		}
	default:
		paths, itemPaths = getAllSubRelPaths(item, pathPrefix)
	}
	return
}

func splitItemPaths(relPath string) []string {
	if relPath == "" {
		return nil
	}
	return strings.Split(relPath, string(filepath.Separator))
}

func getFileItemPath(prefix []string, relPath string) []string {
	return appendCopy(prefix, splitItemPaths(relPath)...)
}

func appendCopy(list []string, incoming ...string) []string {
	m := make([]string, len(list)+len(incoming))
	copy(m, list)
	copy(m[len(list):], incoming)
	return m
}

func getCaseItemPath(prefix []string, relPath string, name string, nameUnderPkg string) []string {
	filePaths := getFileItemPath(prefix, relPath)
	if nameUnderPkg == "" {
		return append(filePaths, name)
	}
	return append(filePaths, strings.Split(nameUnderPkg, "/")...)
}

// emulate the ./pkg/... behavior
func getAllSubRelPaths(t *TestingItem, pathPrefix []string) (relPaths []string, itemPaths [][]string) {
	if t.Kind == TestingItemKind_Case {
		itemPaths = append(itemPaths, getCaseItemPath(pathPrefix, t.RelPath, t.Name, t.NameUnderPkg))
	} else {
		itemPaths = append(itemPaths, getFileItemPath(pathPrefix, t.RelPath))
		if t.HasTestCases && t.Kind != TestingItemKind_File {
			relPaths = append(relPaths, t.RelPath)
		}
	}

	for _, e := range t.Children {
		subDirs, subItemPaths := getAllSubRelPaths(e, pathPrefix)
		relPaths = append(relPaths, subDirs...)
		itemPaths = append(itemPaths, subItemPaths...)
	}
	return
}

func getFileSubCases(t *TestingItem) (arg string, cases []string) {
	arg = filepath.Dir(t.RelPath)
	for _, child := range t.Children {
		if child.Kind != TestingItemKind_Case {
			continue
		}
		cases = append(cases, child.Name)
	}
	return
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

func getPkgSubPath(modPath string, pkgPath string) string {
	// NOTE: pkgPath can be command-line-arguments
	if !strings.HasPrefix(pkgPath, modPath) {
		return ""
	}
	return strings.TrimPrefix(pkgPath[len(modPath):], "/")
}

func makeMergeTree(subPath []string, prefixName string, item *TestingItem) *TestingItem {
	if len(subPath) == 0 {
		return item
	}
	key := subPath[0]
	nameUnderPkg := prefixName + "/" + key
	return &TestingItem{
		Key:            key,
		Name:           key,
		BaseCaseName:   item.BaseCaseName,
		NameUnderPkg:   nameUnderPkg,
		RelPath:        item.RelPath,
		File:           item.File,
		Line:           item.Line,
		Kind:           TestingItemKind_Case,
		HasTestGoFiles: item.HasTestGoFiles,
		HasTestCases:   item.HasTestCases,
		State: &TestingItemState{
			Expanded: true,
		},
		Children: []*TestingItem{makeMergeTree(subPath[1:], nameUnderPkg, item)},
	}
}

func convTestingEvents(events []interface{}) []*TestingItemEvent {
	testingEvents := make([]*TestingItemEvent, 0, len(events))
	for _, e := range events {
		testingEvents = append(testingEvents, e.(*TestingItemEvent))
	}
	return testingEvents
}

func (c *runSession) sendEvent(event *TestingItemEvent) {
	c.session.SendEvents(event)
}

func run(ctx *RunContext) error {
	projectDir := ctx.ProjectDir
	relPath := ctx.RelPath
	name := ctx.Name
	buildFlags := ctx.BuildFlags
	runArgs := ctx.Args
	env := ctx.Env
	goCmd := ctx.GoCmd
	stderr := ctx.Stderr
	stdout := ctx.Stdout
	verbose := true

	args := []string{"test", "-run", fmt.Sprintf("^%s$", name)}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, buildFlags...)
	args = append(args, "./"+filepath.Dir(relPath))
	if len(runArgs) > 0 {
		args = append(args, "-args")
		args = append(args, runArgs...)
	}

	if goCmd == "" {
		goCmd = "go"
	}
	return cmd.Debug().Dir(projectDir).
		Env(env).
		Stdout(stdout).
		Stderr(stderr).
		Run(goCmd, args...)
}

func getTestFlags(absProjectDir string, file string, name string) (flags []string, args []string, err error) {
	_, funcs, err := parseTestFuncs(file)
	if err != nil {
		return nil, nil, err
	}
	fn, err := getFuncDecl(funcs, name)
	if err != nil {
		return nil, nil, err
	}
	flags, args, err = parseFuncArgs(fn)
	if err != nil {
		return nil, nil, err
	}
	return applyVars(absProjectDir, flags), applyVars(absProjectDir, args), nil
}
