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

	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/script/xgo.helper/instrument"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/git"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/osinfo"
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

// run all tests under runtime/test, but not nested go mod
//   go run ./script/run-test --include go1.20.14 --include ./runtime/test/all

// debug:
//   go run ./script/run-test --include go1.20.14 --debug ./runtime/test/patch      # will debug runtime
//   go run ./script/run-test --include go1.20.14 --debug-xgo ./runtime/test/patch  # will debug instrumentation

// when will cache affect?
//   -tags dev : cache is off by default, so revision is not significant
//   otherwise, revision is used as cache key

var globalFlags = []string{"-timeout=60s"}

type TestConfig struct {
	Dir            string
	UsePlainGo     bool
	UsePrebuiltXgo bool
	BuildOnly      bool // implies test -c -o /dev/null

	Go string // minimal go version

	Args            []string
	Flags           []string
	VendorIfMissing bool
	Env             []string
}

func init() {
	// go list -e ./... | grep -Ev 'asset|internal/vendir'
	rootTestsOutput, err := cmd.Output("go", "list", "-e", "./...")
	if err != nil {
		panic(err)
	}
	list := strings.Split(rootTestsOutput, "\n")
	var defaultTestArgs []string
	for _, s := range list {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.Contains(s, "asset") || strings.Contains(s, "internal/vendir") {
			continue
		}
		defaultTestArgs = append(defaultTestArgs, s)
	}
	// fmt.Printf("DEBUG defaultTestArgs: %v\n", defaultTestArgs)

	defaultTest.Args = defaultTestArgs
}

var defaultTest = &TestConfig{
	// test using go
	// go test -v $(go list -e ./... | grep -Ev 'asset|internal/vendir')
	Dir:        "",
	UsePlainGo: true,
	// go list -e ./... | grep -Ev 'asset|internal/vendir'
	Args: nil, // filled later
}
var runtimeTest = &TestConfig{
	Dir:        "runtime",
	UsePlainGo: true,
	Args: []string{
		// "./...",
		"./core/...",
	},
}

// these tests are predefined because
// I don't want to add a test-config.txt
// to them
var predefinedTests = []*TestConfig{
	defaultTest,
	runtimeTest,
	// TODO: check extraSubTests for remaining tests
}

func main() {
	args := os.Args[1:]
	var excludes []string
	var includes []string

	var remainArgs []string
	var remainTests []string

	var resetInstrument bool

	var installXgo bool

	var debug bool
	var debugXgo bool
	var debugGo bool
	var debugCompile string

	var cover bool
	var coverpkgs []string
	var coverprofile string
	var withGoroot string

	var projectDir string

	var logDebug bool
	var withSetup bool
	var withSetupOnly bool
	var tags string
	var flagRace bool

	var flagList bool
	var flagJSON bool
	if len(args) > 0 && args[0] == "list" {
		flagList = true
		args = args[1:]
	}
	n := len(args)
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
		if arg == "--with-setup-only" {
			withSetupOnly = true
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
		if arg == "--with-goroot" {
			if i+1 >= n {
				panic(fmt.Errorf("%v requires arg", arg))
			}
			withGoroot = args[i+1]
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
		if arg == "--debug-go" {
			debugGo = true
			continue
		}
		if arg == "--debug-compile" {
			var val string
			if i+1 < n && !strings.HasPrefix(args[i+1], "-") {
				val = args[i+1]
				i++
			}
			if val == "" {
				val = "true"
			}
			debugCompile = val
			continue
		} else if strings.HasPrefix(arg, "--debug-compile=") {
			val := arg[len("--debug-compile="):]
			if val == "" {
				val = "true"
			}
			debugCompile = val
			continue
		}

		if arg == "--install-xgo" {
			installXgo = true
			continue
		}
		if arg == "--json" {
			flagJSON = true
			continue
		}
		if arg == "-tags" {
			if i+1 >= n {
				panic(fmt.Errorf("%v requires arg", arg))
			}
			tags = args[i+1]
			i++
			continue
		}
		if arg == "-race" {
			flagRace = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				remainArgs = append(remainArgs, arg)
				continue
			}
		}
		fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
		os.Exit(1)
	}
	if flagList {
		list(remainTests, flagJSON, predefinedTests)
		return
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
	if withGoroot == "" {
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
	} else {
		goroots = []string{withGoroot}
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
	if withSetup || withSetupOnly {
		setupGoroots = make([]string, len(goroots))
		for i, goroot := range goroots {
			setupGoroots[i], err = setupGoroot(goroot, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "setup goroot: %s %v\n", goroot, err)
				os.Exit(1)
			}
		}
		if withSetupOnly {
			goroots = setupGoroots
		} else {
			goroots = append(goroots, setupGoroots...)
		}
	}
	var debugNum int
	if debug {
		debugNum++
	}
	if debugXgo {
		debugNum++
	}
	if debugCompile != "" {
		debugNum++
	}
	if debugGo {
		debugNum++
	}
	if debugNum > 0 {
		if debugNum > 1 {
			fmt.Fprintf(os.Stderr, "--debug, --debug-xgo, --debug-go and --debug-compile cannot be used together\n")
			os.Exit(1)
		}
		if len(goroots) > 1 {
			fmt.Fprintf(os.Stderr, "--debug* is not supported with multiple goroots, use --include GOROOT to specify one\n")
			os.Exit(1)
		}
	}

	for _, goroot := range goroots {
		var isSetupGoroot bool
		for _, setupGoroot := range setupGoroots {
			if setupGoroot == goroot {
				isSetupGoroot = true
				break
			}
		}

		begin := time.Now()
		fmt.Fprintf(os.Stdout, "TEST %s\n", goroot)
		if resetInstrument {
			// reset-instrument: reset goroot, but not caches
			if logDebug {
				fmt.Printf("resetting instrument\n")
			}
			cmdArgs := []string{
				"run", "./cmd/xgo", "setup", "--reset-instrument", "--with-goroot", goroot,
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
		goVersion, err := goinfo.GetGorootVersion(goroot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get go version: %v\n", err)
			os.Exit(1)
		}

		if projectDir != "" {
			fmt.Fprintf(os.Stderr, "--project-dir is not supported\n")
			os.Exit(1)
		}
		testArgs := getTests(remainTests, predefinedTests)
		if len(testArgs) == 0 {
			fmt.Fprintf(os.Stderr, "no tests to run: %v\n", remainTests)
			os.Exit(1)
		}
		var goExe string
		if debugNum > 0 {
			if len(testArgs) > 1 {
				fmt.Fprintf(os.Stderr, "--debug* is not supported with multiple tests, please specify exact test\n")
				os.Exit(1)
			}
			if debugCompile != "" {
				err := instrument.InstrumentGc(goroot, goVersion)
				if err != nil {
					fmt.Fprintf(os.Stderr, "instrument gc: %v\n", err)
					os.Exit(1)
				}
				// build
				goExe, err = build.BuildGoDebugBinary(goroot)
				if err != nil {
					fmt.Fprintf(os.Stderr, "build go debug binary: %v\n", err)
					os.Exit(1)
				}
				_, err = build.BuildGoToolCompileDebugBinary(goroot)
				if err != nil {
					fmt.Fprintf(os.Stderr, "build go tool compile: %v\n", err)
					os.Exit(1)
				}
			}
			if debugGo {
				// build
				goExe, err = build.BuildGoDebugBinary(goroot)
				if err != nil {
					fmt.Fprintf(os.Stderr, "build go debug binary: %v\n", err)
					os.Exit(1)
				}
			}
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
			if debugGo && !testArg.UsePlainGo {
				fmt.Fprintf(os.Stderr, "--debug-go must be used with go, not xgo\n")
				os.Exit(1)
			}
			usePlainGo := testArg.UsePlainGo
			useBuiltXgo := testArg.UsePrebuiltXgo
			buildOnly := testArg.BuildOnly

			if testArg.Go != "" {
				testGoVersion, err := goinfo.ParseGoVersionNumber(testArg.Go)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid go version: %v\n", err)
					os.Exit(1)
				}
				if goVersion.Major < testGoVersion.Major || goVersion.Minor < testGoVersion.Minor {
					// skipped
					continue
				}
			}
			// projectDir
			runArgs := make([]string, 0, len(remainArgs)+1)
			opts := Opts{
				Tags:           tags,
				IsSetupGoroot:  isSetupGoroot,
				UsePrebuiltXgo: useBuiltXgo,
				BuildOnly:      buildOnly,
				GoBinary:       goExe,
				DebugGo:        debugGo,
			}
			if tags != "" {
				runArgs = append(runArgs, "-tags", tags)
				opts.Tags = tags
			}
			if !usePlainGo {
				if logDebug {
					runArgs = append(runArgs, "--log-debug")
				}
				if debug {
					runArgs = append(runArgs, "--debug")
				}
				if debugXgo {
					opts.DebugXgo = true
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
			if testArg.VendorIfMissing {
				if !hasSubDir(testArg.Dir, "vendor") {
					fmt.Fprintf(os.Stderr, "cd %s && go mod vendor\n", testArg.Dir)
					err := cmd.Dir(testArg.Dir).Run("go", "mod", "vendor")
					if err != nil {
						fmt.Fprintf(os.Stderr, "vendor: %v\n", err)
						os.Exit(1)
					}
				}
			}

			env := testArg.Env
			if debugCompile != "" {
				listArg := []string{debugCompile}
				if debugCompile == "true" {
					listArg = testArg.Args
				}
				// list pkg
				listArgs := []string{"list"}
				listArgs = append(listArgs, listArg...)
				// fmt.Fprintf(os.Stderr, "go %v\n", listArgs)
				output, err := cmd.Dir(testArg.Dir).Output("go", listArgs...)
				if err != nil {
					fmt.Fprintf(os.Stderr, "list pkg: %v\n", err)
					os.Exit(1)
				}
				pkgs := strings.Split(string(output), "\n")
				if len(pkgs) == 0 {
					fmt.Fprintf(os.Stderr, "no pkg found\n")
					os.Exit(1)
				}
				if len(pkgs) > 1 {
					fmt.Fprintf(os.Stderr, "multiple pkgs found, please specify exact --debug-compile=pkg\n")
					os.Exit(1)
				}
				pkg := pkgs[0]
				env = append(env, instrument.XGO_HELPER_DEBUG_PKG+"="+pkg)
			}
			runArgs = addGoFlags(runArgs, cover, coverpkgs, coverprofile, coverageVariant)
			if flagRace {
				runArgs = append(runArgs, "-race")
			}
			if debugCompile != "" {
				runArgs = append(runArgs, "-a")
			}
			runArgs = append(runArgs, testArg.Flags...)
			runArgs = append(runArgs, remainArgs...)

			err := doRunTest(goroot, usePlainGo, dir, runArgs, testArg.Args, env, opts)
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
	fmt.Printf("go %s\n", strings.Join(args, " "))
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

func hasSubDir(dir string, subDir string) bool {
	subDirPath := filepath.Join(dir, subDir)
	_, err := os.Stat(subDirPath)
	return err == nil
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

type testKind string

type Opts struct {
	Tags           string
	IsSetupGoroot  bool
	UsePrebuiltXgo bool
	BuildOnly      bool
	GoBinary       string
	DebugXgo       bool
	DebugGo        bool
}

func doRunTest(goroot string, usePlainGo bool, dir string, args []string, tests []string, env []string, opts ...Opts) error {
	var debugXgo bool
	var tags string
	var isSetupGoroot bool
	var useBuiltXgo bool
	var buildOnly bool
	var goBinary string
	var debugGo bool
	if len(opts) > 0 {
		if len(opts) != 1 {
			panic("only one opts is allowed")
		}
		opt := opts[0]
		debugXgo = opt.DebugXgo
		tags = opt.Tags
		isSetupGoroot = opt.IsSetupGoroot
		useBuiltXgo = opt.UsePrebuiltXgo
		buildOnly = opt.BuildOnly
		goBinary = opt.GoBinary
		debugGo = opt.DebugGo
	}
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}
	var prebuiltXgo string
	var testArgs []string
	if usePlainGo {
		testArgs = []string{"test"}
	} else if useBuiltXgo {
		buildArgs := []string{"build"}
		if tags != "" {
			buildArgs = append(buildArgs, "-tags", tags)
		}

		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		prebuiltXgo = filepath.Join(wd, "__debug_bin_xgo"+osinfo.EXE_SUFFIX)
		buildArgs = append(buildArgs, "-o", prebuiltXgo, "./cmd/xgo")
		err = cmd.Run("go", buildArgs...)
		if err != nil {
			return err
		}
		testArgs = []string{"test"}
	} else {
		if !debugXgo {
			testArgs = []string{"run"}
			if tags != "" {
				testArgs = append(testArgs, "-tags", tags)
			}
			testArgs = append(testArgs, "./cmd/xgo", "test")
		} else {
			testArgs = []string{"run"}
			if tags != "" {
				testArgs = append(testArgs, "-tags", tags)
			}
			testArgs = append(testArgs, "./cmd/xgo", "test")
		}
	}
	if buildOnly {
		testArgs = append(testArgs, "-c", "-o", "/dev/null")
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

	for _, env := range env {
		fmt.Fprintln(os.Stderr, env)
	}

	// debug
	var binary string
	if debugXgo {
		fmt.Printf("testArgs: %v\n", testArgs)
		binary = "xgo"
		newTestArgs := make([]string, 2, len(testArgs)+1)
		newTestArgs[0] = testArgs[0]
		newTestArgs[1] = "--debug"
		newTestArgs = append(newTestArgs, testArgs[1:]...)
		testArgs = newTestArgs
		fmt.Printf("xgo %s\n", strings.Join(testArgs, " "))
	} else if useBuiltXgo {
		binary = prebuiltXgo
		fmt.Printf("./xgo %s\n", strings.Join(testArgs, " "))
	} else {
		if goBinary != "" {
			binary = goBinary
		} else {
			binary = filepath.Join(goroot, "bin", "go")
		}
		fmt.Printf("%s %s\n", binary, strings.Join(testArgs, " "))
	}
	if debugGo {
		if goBinary == "" {
			return fmt.Errorf("--debug-go requires go binary to be specified")
		}
		vscodeDebug := strings.ReplaceAll(instrument.VScodeDebug, "2346", "2345")
		fmt.Fprintf(os.Stderr, "VSCode add the following config to .vscode/launch.json:\n%s\nThen set breakpoint at %s\n", vscodeDebug, filepath.Join(goroot, "src", "cmd", "go", "main.go"))
		testArgs = append([]string{"exec", "--listen=:2345", "--api-version=2", "--check-go-version=false", "--headless", "--", goBinary}, testArgs...)
		binary = "dlv"
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
	if isSetupGoroot {
		execCmd.Env = append(execCmd.Env, "XGO_IS_SETUP_GOROOT=true")
	}
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
