package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/git"
)

// the test driver for xgo
// it automatically detects all GOROOTs under `go-release`, and
// run predefined tests for each GOROOT.
// this simplifies xgo's compatibility test across multiple go versions.
// specifically, `go-release` directory can be downloaded using:
//   go run ./script/download-go go1.19.13
// if there is no any go version under `go-release`, the system GOROOT will be used.

// usage:
//
//	go run ./script/run-test/ --include go1.19.13
//	go run ./script/run-test/ -count=1 --include go1.19.13
//	go run ./script/run-test/ --include go1.19.13 -run TestHelloWorld -v
//	go run ./script/run-test/ -count=1 --include go1.17.13 --include go1.18.10 --include go1.19.13 --include go1.20.14 --include go1.21.8 --include go1.22.2
//  go run ./script/run-test/ -cover -coverpkg github.com/xhd2015/xgo/runtime/... -coverprofile covers/cover.out --include go1.21.8

// specific test:
//   go run ./script/run-test --include go1.18.10 --log-debug ./runtime/test/patch

// debug:
//   go run ./script/run-test --include go1.20.14 --debug ./runtime/test/patch

// when will cache affect?
//   -tags dev : cache is off by default, so revision is not significant
//   otherwise, revision is used as cache key

var globalFlags = []string{"-timeout=60s"}

type TestArg struct {
	Dir  string
	Args []string
}

var defaultTestArgs = []*TestArg{
	{
		// test using go
		Dir: "",
		Args: []string{
			// "./...",
			// "./test",
			"./support/...",
		},
	},
	{
		// test using go
		Dir: "runtime",
		Args: []string{
			// "./...",
			"./core/...",
		},
	},
	{
		// test using xgo
		Dir: "runtime/test",
		Args: []string{
			// "./...",
			"./functab",
			"./patch",
			"./mock/mock_by_name",
			"./mock/mock_func",
			"./mock/mock_generic",
			"./mock/mock_method",
			"./mock/mock_var",
			"./trace/record",
			"./trace/go_trace",
			"./trace/marshal/cyclic",
			"./trap/inspect",
			"./trap/interceptor",
			"./tls",
		},
	},
}

func main() {
	args := os.Args[1:]
	var excludes []string
	var includes []string

	n := len(args)
	var remainArgs []string
	var remainTests []string

	var resetInstrument bool

	var installXgo bool

	var debug bool
	var debugXgo bool
	var cover bool
	var coverpkgs []string
	var coverprofile string

	var projectDir string

	var logDebug bool
	var withSetup bool

	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			if i+1 < n {
				remainArgs = append(remainArgs, args[i+1:]...)
			}
			break
		}
		if !strings.HasPrefix(arg, "-") {
			remainTests = append(remainTests, arg)
			continue
		}
		if arg == "--exclude" {
			excludes = append(excludes, args[i+1])
			i++
			continue
		}
		if arg == "--include" {
			includes = append(includes, args[i+1])
			i++
			continue
		}
		if arg == "--project-dir" {
			if i+1 >= n {
				panic(fmt.Errorf("%v requires arg", arg))
			}
			projectDir = args[i+1]
			i++
			continue
		}
		if arg == "-run" {
			remainArgs = append(remainArgs, arg, args[i+1])
			i++
			continue
		}
		if arg == "--reset-instrument" {
			resetInstrument = true
			continue
		}
		if arg == "-a" {
			remainArgs = append(remainArgs, arg)
			continue
		}
		if arg == "-v" || strings.HasPrefix(arg, "-count=") {
			remainArgs = append(remainArgs, arg)
			continue
		}
		if arg == "--log-debug" {
			logDebug = true
			continue
		}
		if arg == "--with-setup" {
			withSetup = true
			continue
		}
		if arg == "-cover" {
			cover = true
			continue
		}
		if arg == "-coverpkg" {
			coverpkgs = append(coverpkgs, args[i+1])
			i++
			continue
		}
		if arg == "-coverprofile" {
			coverprofile = args[i+1]
			i++
			continue
		}
		if arg == "--debug" {
			debug = true
			continue
		}
		if arg == "--debug-xgo" {
			debugXgo = true
			continue
		}
		if arg == "--install-xgo" {
			installXgo = true
			continue
		}
		fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
		os.Exit(1)
	}
	_ = projectDir
	topLevel, err := git.ShowTopLevel("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	same := func(a string, b string) (bool, error) {
		absA, err := filepath.Abs(a)
		if err != nil {
			return false, err
		}
		absB, err := filepath.Abs(b)
		if err != nil {
			return false, err
		}
		return absA == absB, nil
	}
	atRoot, err := same(topLevel, wd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !atRoot {
		fmt.Fprintf(os.Stderr, "run-test requires executing from project root, current: %s\n", wd)
		os.Exit(1)
	}
	if coverprofile != "" {
		absProfile, err := filepath.Abs(coverprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		coverprofile = absProfile
	}
	goRelease := "go-release"
	var goroots []string
	_, statErr := os.Stat(goRelease)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			fmt.Fprintf(os.Stderr, "stat: %v\n", statErr)
			os.Exit(1)
			return
		}
		// use default GOROOT
		goroot := runtime.GOROOT()
		if goroot == "" {
			var err error
			goroot, err = cmd.Output("go", "env", "GOROOT")
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot get GOROOT: %v\n", err)
				os.Exit(1)
				return
			}
		}
		goroots = append(goroots, goroot)
	} else {
		goRootNames, err := listGoroots(goRelease)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if len(includes) > 0 && len(excludes) > 0 {
			fmt.Fprintf(os.Stderr, "--exclude and --include cannot be used together\n")
			os.Exit(1)
		}
		if len(includes) > 0 {
			i := 0
			for _, goroot := range goRootNames {
				if listContains(includes, goroot) {
					goRootNames[i] = goroot
					i++
				}
			}
			goRootNames = goRootNames[:i]
		} else if len(excludes) > 0 {
			i := 0
			for _, goroot := range goRootNames {
				if !listContains(excludes, goroot) {
					goRootNames[i] = goroot
					i++
				}
			}
			goRootNames = goRootNames[:i]
		}
		for _, goRootName := range goRootNames {
			goroots = append(goroots, filepath.Join(goRelease, goRootName))
		}
	}

	if len(goroots) == 0 {
		fmt.Fprintf(os.Stderr, "no go select\n")
		os.Exit(1)
	}

	if installXgo {
		err := cmd.Run("go", "install", "./cmd/xgo")
		if err != nil {
			fmt.Fprintf(os.Stderr, "install xgo: %v\n", err)
			os.Exit(1)
		}
	}
	var setupGoroots []string
	if withSetup {
		setupGoroots = make([]string, len(goroots))
		for i, goroot := range goroots {
			setupGoroots[i], err = setupGoroot(goroot, logDebug)
			if err != nil {
				fmt.Fprintf(os.Stderr, "setup goroot: %s %v\n", goroot, err)
				os.Exit(1)
			}
		}
		goroots = append(goroots, setupGoroots...)
	}
	for _, goroot := range goroots {
		begin := time.Now()
		fmt.Fprintf(os.Stdout, "TEST %s\n", goroot)
		if resetInstrument {
			// reset-instrument: reset goroot, but not caches
			if logDebug {
				fmt.Printf("resetting instrument\n")
			}
			cmdArgs := []string{
				"run", "./cmd/xgo", "build", "--reset-instrument", "--with-goroot", goroot,
			}
			if logDebug {
				cmdArgs = append(cmdArgs, "--log-debug=stdout")
			}
			err := cmd.Run("go", cmdArgs...)
			if err != nil {
				if extErr, ok := err.(*exec.ExitError); ok {
					fmt.Fprintf(os.Stderr, "%s\n", extErr.Stderr)
				} else {
					fmt.Fprintf(os.Stderr, "%v", err)
				}
				os.Exit(1)
			}
		}

		if projectDir != "" {
			fmt.Fprintf(os.Stderr, "--project-dir is not supported\n")
			os.Exit(1)
		}
		testArgs := getTestArgs(remainTests)
		if len(testArgs) == 0 {
			fmt.Fprintf(os.Stderr, "no tests to run: %v\n", remainTests)
			os.Exit(1)
		}
		for _, testArg := range testArgs {
			if len(testArg.Args) == 0 {
				logDir := testArg.Dir
				if logDir == "" {
					logDir = "."
				}
				fmt.Fprintf(os.Stderr, "SKIP %s: no tests to run\n", logDir)
				continue
			}
			usePlainGo := testArg.Dir == "" || testArg.Dir == "runtime"
			// projectDir
			runArgs := make([]string, 0, len(remainArgs)+1)
			var opts Opts
			if !usePlainGo {
				if logDebug {
					runArgs = append(runArgs, "--log-debug")
				}
				if debug {
					runArgs = append(runArgs, "--debug")
				}
				if debugXgo {
					opts.Debug = true
				}
			}
			var coverageVariant string
			var dir string
			if testArg.Dir != "" {
				// runtime, runtime-test
				coverageVariant = strings.ReplaceAll(strings.ReplaceAll(testArg.Dir, "/", "-"), "\\", "-")
				if usePlainGo {
					dir = testArg.Dir
				} else {
					runArgs = append(runArgs, "--project-dir", testArg.Dir)
				}
			}
			runArgs = addGoFlags(runArgs, cover, coverpkgs, coverprofile, coverageVariant)

			runArgs = append(runArgs, remainArgs...)
			err := doRunTest(goroot, usePlainGo, dir, runArgs, testArg.Args, nil, opts)
			if err != nil {
				fmt.Fprintf(os.Stdout, "FAIL %s: %v(%v)\n", goroot, err, time.Since(begin))
				os.Exit(1)
			}
		}

		fmt.Fprintf(os.Stdout, "PASS %s(%v)\n", goroot, time.Since(begin))
	}
}

func setupGoroot(goroot string, logDebug bool) (string, error) {
	args := []string{
		"run", "./cmd/xgo", "setup", "--with-goroot", goroot,
	}
	if logDebug {
		args = append(args, "--log-debug")
	}
	return cmd.Output("go", args...)
}

func addGoFlags(args []string, cover bool, coverPkgs []string, coverprofile string, coverageVariant string) []string {
	if cover {
		args = append(args, "-cover")
	}
	for _, coverPkg := range coverPkgs {
		args = append(args, "-coverpkg", coverPkg)
	}
	if coverprofile != "" {
		var prefix string
		var suffix string
		idx := strings.LastIndex(coverprofile, ".")
		if idx < 0 {
			prefix = coverprofile
		} else {
			prefix, suffix = coverprofile[:idx], coverprofile[idx:]
		}
		args = append(args, "-coverprofile", prefix+"-"+coverageVariant+suffix)
	}
	return args
}

func getTestArgs(args []string) []*TestArg {
	if len(args) == 0 {
		return defaultTestArgs
	}
	return splitArgs(args)
}

func splitArgs(args []string) []*TestArg {
	var results []*TestArg
	for _, arg := range args {
		dir, path := splitArg(arg)

		var prev *TestArg
		for i, r := range results {
			if r.Dir == dir {
				prev = results[i]
				break
			}
		}
		if prev == nil {
			prev = &TestArg{Dir: dir, Args: []string{}}
			results = append(results, prev)
		}
		prev.Args = append(prev.Args, path)
	}
	return results
}

func splitArg(fullArg string) (dir string, path string) {
	if fullArg == "" {
		return "", ""
	}
	parts := splitPath(fullArg)
	if len(parts) == 0 {
		return "", ""
	}
	if parts[0] != "." {
		xgo := []string{"github.com", "xhd2015", "xgo"}
		if !isList(parts, xgo) {
			return "", fullArg
		}
		return splitXgoRelative(parts[len(xgo):], fullArg)
	}
	return splitXgoRelative(parts[1:], fullArg)
}

func splitXgoRelative(relative []string, fullArg string) (string, string) {
	if len(relative) == 0 {
		return "", fullArg
	}
	if relative[0] != "runtime" {
		return "", fullArg
	}
	if len(relative) >= 2 && relative[1] == "test" {
		return filepath.Join("runtime", "test"), "./" + strings.Join(relative[2:], "/")
	}
	return "runtime", "./" + strings.Join(relative[1:], "/")
}

func isList(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if e != b[i] {
			return false
		}
	}
	return true
}
func splitPath(path string) []string {
	var list []string
	var runes []rune
	for _, r := range path {
		if r != '/' && r != '\\' {
			runes = append(runes, r)
			continue
		}
		if len(runes) > 0 {
			list = append(list, string(runes))
			runes = nil
		}
	}
	if len(runes) > 0 {
		list = append(list, string(runes))
	}
	return list
}

// TODO: use slices.Contains()
func listContains(list []string, s string) bool {
	for _, e := range list {
		if s == e {
			return true
		}
	}
	return false
}
func listGoroots(dir string) ([]string, error) {
	subDirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			continue
		}
		if !strings.HasPrefix(subDir.Name(), "go") {
			continue
		}
		dirs = append(dirs, subDir.Name())
	}
	return dirs, nil
}

func hasName(names []string, name string) bool {
	if len(names) == 0 {
		return true
	}
	for _, e := range names {
		if e == name {
			return true
		}
	}
	return false
}

func runRuntimeSubTest(goroot string, args []string, tests []string, names []string) error {
	amendArgs := func(customArgs []string, flags []string) []string {
		newArgs := make([]string, 0, len(args)+len(customArgs)+len(flags))
		newArgs = append(newArgs, args...)
		newArgs = append(newArgs, customArgs...)
		newArgs = append(newArgs, flags...)
		return newArgs
	}
	for _, tt := range extraSubTests {
		if !hasName(names, tt.name) {
			continue
		}
		if tt.skipOnCover {
			var hasCoverFlag bool
			for _, arg := range args {
				if strings.HasPrefix(arg, "-cover") {
					hasCoverFlag = true
					break
				}
			}
			if hasCoverFlag {
				fmt.Printf("skip on cover: %s\n", tt.name)
				continue
			}
		}
		dir := tt.dir
		env := tt.env
		var subDirs bool
		if strings.HasSuffix(dir, "/...") {
			subDirs = true
			dir = strings.TrimSuffix(dir, "/...")
		}
		var hasHook bool
		if !subDirs {
			hookFile := filepath.Join(dir, "hook_test.go")
			_, statErr := os.Stat(hookFile)
			hasHook = statErr == nil
		}
		dotDir := "./" + dir

		var runDir string
		var extraArgs []string
		if tt.usePlainGo {
			runDir = dotDir
		} else {
			extraArgs = []string{"--project-dir", dotDir}
		}

		if hasHook {
			err := doRunTest(goroot, tt.usePlainGo, runDir, amendArgs(append(extraArgs, []string{"-run", "TestPreCheck"}...), nil), []string{"./hook_test.go"}, env)
			if err != nil {
				return err
			}
		}
		var testArgDir string
		if !subDirs {
			testArgDir = "./"
		} else {
			testArgDir = "./..."
		}
		testFlags := tt.flags
		if runtime.GOOS == "windows" && tt.windowsFlags != nil {
			testFlags = tt.windowsFlags
		}
		var skipped bool
		runErr := doRunTest(goroot, tt.usePlainGo, runDir, amendArgs(extraArgs, testFlags), []string{testArgDir}, env)
		if runErr != nil {
			if tt.windowsFailIgnore && runtime.GOOS == "windows" {
				skipped = true
				fmt.Printf("SKIP test failure on windows\n")
			}
			if !skipped && tt.skipOnTimeout {
				// see https://github.com/xhd2015/xgo/issues/272#issuecomment-2466539209
				if runErr, ok := runErr.(*commandError); ok && runErr.timeoutDetector.found() {
					skipped = true
					fmt.Printf("SKIP test timed out\n")
				}
			}
			if !skipped {
				return runErr
			}
		}

		if !skipped && hasHook {
			err := doRunTest(goroot, tt.usePlainGo, runDir, amendArgs(append(extraArgs, []string{"-run", "TestPostCheck"}...), nil), []string{"./hook_test.go"}, env)
			if err != nil {
				return err
			}
		}
	}
	if len(names) == 0 {
		panic("TODO fix")
		err := doRunTest(goroot, false, "", args, tests, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

type testKind string

type Opts struct {
	Debug bool
}

func doRunTest(goroot string, usePlainGo bool, dir string, args []string, tests []string, env []string, opts ...Opts) error {
	var debug bool
	if len(opts) > 0 {
		if len(opts) != 1 {
			panic("only one opts is allowed")
		}
		opt := opts[0]
		debug = opt.Debug
	}
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}
	var testArgs []string
	if !usePlainGo {
		if !debug {
			testArgs = []string{"run", "-tags", "dev", "./cmd/xgo", "test"}
		} else {
			testArgs = []string{"run", "./cmd/xgo", "test"}
		}
	} else {
		testArgs = []string{"test"}
	}

	testArgs = append(testArgs, globalFlags...)
	testArgs = append(testArgs, args...)
	testArgs = append(testArgs, tests...)

	// remove extra xgo flags
	if usePlainGo {
		n := len(testArgs)
		i := 0
		for j := 0; j < n; j++ {
			if testArgs[j] == "--log-debug" || strings.HasPrefix(testArgs[j], "--log-debug=") {
				continue
			}
			testArgs[i] = testArgs[j]
			i++
		}
		testArgs = testArgs[:i]
	}
	// debug

	var binary string
	if debug {
		fmt.Printf("testArgs: %v\n", testArgs)
		binary = "xgo"
		newTestArgs := make([]string, 2, len(testArgs)+1)
		newTestArgs[0] = testArgs[0]
		newTestArgs[1] = "--debug"
		newTestArgs = append(newTestArgs, testArgs[1:]...)
		testArgs = newTestArgs
		fmt.Printf("xgo %s\n", strings.Join(testArgs, " "))
	} else {
		binary = filepath.Join(goroot, "bin", "go")
		fmt.Printf("go %s\n", strings.Join(testArgs, " "))
	}

	execCmd := exec.Command(binary, testArgs...)
	dt := detector{
		match: []byte("panic: test timed out after"),
	}
	execCmd.Stdout = &dt
	execCmd.Stderr = os.Stderr
	execCmd.Dir = dir

	execCmd.Env = os.Environ()
	execCmd.Env = append(execCmd.Env, "GOROOT="+goroot)
	execCmd.Env = append(execCmd.Env, "PATH="+filepath.Join(goroot, "bin")+string(filepath.ListSeparator)+os.Getenv("PATH"))
	execCmd.Env = append(execCmd.Env, env...)

	cmdErr := execCmd.Run()
	if cmdErr != nil {
		return &commandError{err: cmdErr, timeoutDetector: &dt}
	}
	return nil
}

type commandError struct {
	err             error
	timeoutDetector *detector
}

func (c *commandError) Error() string {
	return c.err.Error()
}

type detector struct {
	foundMatch bool
	match      []byte
}

func (d *detector) Write(p []byte) (n int, err error) {
	n = len(p)
	if d.foundMatch {
		return
	}
	if bytes.HasPrefix(p, d.match) {
		d.foundMatch = true
		return
	}
	return os.Stdout.Write(p)
}
func (d *detector) found() bool {
	return d.foundMatch
}
