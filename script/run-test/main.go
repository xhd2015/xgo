package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
)

// usage:
//
//	go run ./script/run-test/ --include go1.19.13
//	go run ./script/run-test/ -count=1 --include go1.19.13
//	go run ./script/run-test/ --include go1.19.13 -run TestHelloWorld -v
//	go run ./script/run-test/ -count=1 --include go1.17.13 --include go1.18.10 --include go1.19.13 --include go1.20.14 --include go1.21.8 --include go1.22.2
//  go run ./script/run-test/ -cover -coverpkg github.com/xhd2015/xgo/runtime/... -coverprofile covers/cover.out --include go1.21.8

// when will cache affect?
//   -tags dev : cache is off by default, so revision is not significant
//   otherwise, revision is used as cache key
//
// runtime test:
//    go run ./script/run-test/ --include go1.17.13 --xgo-runtime-test-only -run TestFuncList -v ./test/func_list
//
// runtime sub test:
//    go run ./script/run-test/ --include go1.17.13 --xgo-runtime-sub-test-only -run TestFuncList -v ./func_list

// runtime sub tests by names:
//    go run ./script/run-test/ --include go1.17.13 --name trace_without_dep --name trap_with_overlay

// xgo default test:
//    go run ./script/run-test/ --include go1.18.10 --xgo-default-test-only -run TestAtomicGenericPtr -v ./test

// xgo test
//    go run ./script/run-test/ --include go1.18.10 --xgo-test-only -run TestFuncNames -v ./test/xgo_test/func_names

// run specific test for all go versions
//     go run ./script/run-test/ --include go1.17.13 --include go1.18.10 --include go1.19.13 --include go1.20.14 --include go1.21.8 --include go1.22.1 -count=1 --xgo-default-test-only -run TestFuncNameSliceShouldWorkWithDebug -v

var globalFlags = []string{"-timeout=60s"}

// TODO: remove duplicate test between xgo test and runtime test
var runtimeSubTests = []string{
	"func_list",
	"trap",
	"trap_inspect_func",
	"trap_args",
	"mock_func",
	"mock_method",
	"mock_by_name",
	"mock_closure",
	"mock_stdlib",
	"mock_generic",
	"mock_var",
	"patch",
	"patch_const",
	"tls",
}

type TestCase struct {
	name         string
	usePlainGo   bool //  use go instead of xgo
	dir          string
	flags        []string
	windowsFlags []string
}

var extraSubTests = []*TestCase{
	{
		name:  "trace_without_dep",
		dir:   "runtime/test/trace_without_dep",
		flags: []string{"--strace"},
		// see https://github.com/xhd2015/xgo/issues/144#issuecomment-2138565532
		windowsFlags: []string{"--trap-stdlib=false", "--strace"},
	},
	{
		name:         "trace_without_dep_vendor",
		dir:          "runtime/test/trace_without_dep_vendor",
		flags:        []string{"--strace"},
		windowsFlags: []string{"--trap-stdlib=false", "--strace"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/87
		name:         "trace_without_dep_vendor_replace",
		dir:          "runtime/test/trace_without_dep_vendor_replace",
		flags:        []string{"--strace"},
		windowsFlags: []string{"--trap-stdlib=false", "--strace"},
	},
	{
		name:  "trap_with_overlay",
		dir:   "runtime/test/trap_with_overlay",
		flags: []string{"-overlay", "overlay.json"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/111
		name:  "trap_stdlib_any",
		dir:   "runtime/test/trap_stdlib_any",
		flags: []string{"--trap-stdlib"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/142
		name:  "bugs_regression",
		dir:   "runtime/test/bugs/...",
		flags: []string{},
	},
	{
		name:         "trace_marshal",
		dir:          "runtime/test/trace_marshal/...",
		flags:        []string{},
		windowsFlags: []string{"--trap-stdlib=false"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/142
		name:         "trace",
		dir:          "runtime/test/trace/...",
		flags:        []string{},
		windowsFlags: []string{"--trap-stdlib=false"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/202
		name: "asm_func",
		dir:  "runtime/test/issue_194_asm_func",
	},
	{
		// see https://github.com/xhd2015/xgo/issues/194
		name: "asm_func_sonic",
		dir:  "runtime/test/issue_194_asm_func/demo",
	},
	{
		// see https://github.com/xhd2015/xgo/issues/142
		name:         "trace_panic_peek",
		dir:          "runtime/test/trace_panic_peek/...",
		flags:        []string{},
		windowsFlags: []string{"--trap-stdlib=false"},
	},
	{
		// see https://github.com/xhd2015/xgo/issues/164
		name:  "stdlib_recover_no_trap",
		dir:   "runtime/test/recover_no_trap",
		flags: []string{"--trap-stdlib"},
	},
	{
		name:       "xgo_integration",
		usePlainGo: true,
		dir:        "test/xgo_integration",
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
	var xgoTestOnly bool
	var xgoRuntimeTestOnly bool
	var xgoRuntimeSubTestOnly bool
	var xgoDefaultTestOnly bool

	var debug bool
	var cover bool
	var coverpkgs []string
	var coverprofile string
	var names []string
	for i := 0; i < n; i++ {
		arg := args[i]
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
		if arg == "--xgo-test-only" {
			xgoTestOnly = true
			xgoRuntimeTestOnly = false
			xgoRuntimeSubTestOnly = false
			xgoDefaultTestOnly = false
			continue
		}
		if arg == "--xgo-runtime-test-only" {
			xgoTestOnly = false
			xgoRuntimeTestOnly = true
			xgoRuntimeSubTestOnly = false
			xgoDefaultTestOnly = false
			continue
		}
		if arg == "--xgo-runtime-sub-test-only" {
			xgoTestOnly = false
			xgoRuntimeTestOnly = false
			xgoRuntimeSubTestOnly = true
			xgoDefaultTestOnly = false
			continue
		}
		if arg == "--xgo-default-test-only" {
			xgoTestOnly = false
			xgoRuntimeTestOnly = false
			xgoRuntimeSubTestOnly = false
			xgoDefaultTestOnly = true
			continue
		}
		if arg == "--name" {
			names = append(names, args[i+1])
			i++
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
		fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
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
	_, err := os.Stat(goRelease)
	if err != nil {
		// use default GOROOT
		goroot := runtime.GOROOT()
		if goroot == "" {
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
	for _, goroot := range goroots {
		begin := time.Now()
		fmt.Fprintf(os.Stdout, "TEST %s\n", goroot)
		if resetInstrument {
			// reset-instrument: reset goroot, but not caches
			if debug {
				fmt.Printf("resetting instrument\n")
			}
			cmdArgs := []string{
				"run", "./cmd/xgo", "build", "--reset-instrument", "--with-goroot", goroot, "--build-compiler",
			}
			if debug {
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
		runDefault := true
		runXgoTestFlag := true
		runRuntimeTestFlag := true
		runRuntimeSubTestFlag := true
		if len(names) > 0 {
			runRuntimeTestFlag = false
			runRuntimeSubTestFlag = true
			runXgoTestFlag = false
			runDefault = false
		} else if xgoTestOnly {
			runRuntimeTestFlag = false
			runRuntimeSubTestFlag = false
			runXgoTestFlag = true
			runDefault = false
		} else if xgoRuntimeTestOnly {
			runRuntimeTestFlag = true
			runRuntimeSubTestFlag = false
			runXgoTestFlag = false
			runDefault = false
		} else if xgoRuntimeSubTestOnly {
			runRuntimeTestFlag = false
			runRuntimeSubTestFlag = true
			runXgoTestFlag = false
			runDefault = false
		} else if xgoDefaultTestOnly {
			runRuntimeTestFlag = false
			runRuntimeSubTestFlag = false
			runXgoTestFlag = false
			runDefault = true
		}
		addArgs := func(args []string, variant string) []string {
			if cover {
				args = append(args, "-cover")
			}
			for _, coverPkg := range coverpkgs {
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
				args = append(args, "-coverprofile", prefix+"-"+variant+suffix)
			}
			if debug && (variant == "xgo" || variant == "runtime" || variant == "runtime-sub") {
				args = append(args, "--log-debug=stdout")
			}
			return args
		}
		if runXgoTestFlag {
			err := runXgoTest(goroot, addArgs(remainArgs, "xgo"), remainTests)
			if err != nil {
				fmt.Fprintf(os.Stdout, "FAIL %s: %v(%v)\n", goroot, err, time.Since(begin))
				os.Exit(1)
			}
		}

		if runRuntimeTestFlag {
			err := runRuntimeTest(goroot, addArgs(remainArgs, "runtime"), remainTests)
			if err != nil {
				fmt.Fprintf(os.Stdout, "FAIL %s: %v(%v)\n", goroot, err, time.Since(begin))
				os.Exit(1)
			}
		}
		if runRuntimeSubTestFlag {
			err := runRuntimeSubTest(goroot, addArgs(remainArgs, "runtime-sub"), remainTests, names)
			if err != nil {
				fmt.Fprintf(os.Stdout, "FAIL %s: %v(%v)\n", goroot, err, time.Since(begin))
				os.Exit(1)
			}
		}
		if runDefault {
			err := runDefaultTest(goroot, addArgs(remainArgs, "default"), remainTests)
			if err != nil {
				fmt.Fprintf(os.Stdout, "FAIL %s: %v(%v)\n", goroot, err, time.Since(begin))
				os.Exit(1)
			}
		}
		fmt.Fprintf(os.Stdout, "PASS %s(%v)\n", goroot, time.Since(begin))
	}
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

func runDefaultTest(goroot string, args []string, tests []string) error {
	return doRunTest(goroot, testKind_default, false, "", args, tests)
}

func runXgoTest(goroot string, args []string, tests []string) error {
	return doRunTest(goroot, testKind_xgoTest, false, "", args, tests)
}

func runRuntimeTest(goroot string, args []string, tests []string) error {
	err := doRunTest(goroot, testKind_runtimeTest, false, "", args, tests)
	if err != nil {
		return err
	}
	return nil
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
		dir := tt.dir
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
			err := doRunTest(goroot, testKind_xgoAny, tt.usePlainGo, runDir, amendArgs(append(extraArgs, []string{"-run", "TestPreCheck"}...), nil), []string{"./hook_test.go"})
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
		err := doRunTest(goroot, testKind_xgoAny, tt.usePlainGo, runDir, amendArgs(extraArgs, testFlags), []string{testArgDir})
		if err != nil {
			return err
		}

		if hasHook {
			err := doRunTest(goroot, testKind_xgoAny, tt.usePlainGo, runDir, amendArgs(append(extraArgs, []string{"-run", "TestPostCheck"}...), nil), []string{"./hook_test.go"})
			if err != nil {
				return err
			}
		}
	}
	if len(names) == 0 {
		err := doRunTest(goroot, testKind_runtimeSubTest, false, "", args, tests)
		if err != nil {
			return err
		}
	}
	return nil
}

type testKind string

const (
	testKind_default        testKind = ""
	testKind_xgoTest        testKind = "xgo-test"
	testKind_runtimeTest    testKind = "runtime-test"
	testKind_runtimeSubTest testKind = "runtime-sub-test"
	testKind_xgoAny         testKind = "xgo-any"
)

func doRunTest(goroot string, kind testKind, usePlainGo bool, dir string, args []string, tests []string) error {
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}
	var testArgs []string
	switch kind {
	case testKind_default:
		testArgs = []string{"test"}
		testArgs = append(testArgs, globalFlags...)
		testArgs = append(testArgs, args...)
		if len(tests) > 0 {
			testArgs = append(testArgs, tests...)
		} else {
			// TODO: enable integration tests on windows
			if runtime.GOOS != "windows" {
				testArgs = append(testArgs, "./test")
			}
			testArgs = append(testArgs, "./support/...")
			// exclude ./cmd/xgo/runtime_gen
			testArgs = append(testArgs, "./cmd/xgo")
			testArgs = append(testArgs, "./cmd/xgo/exec_tool/...")
			testArgs = append(testArgs, "./cmd/xgo/patch/...")
			testArgs = append(testArgs, "./cmd/xgo/pathsum/...")
			testArgs = append(testArgs, "./cmd/xgo/trace/...")
			testArgs = append(testArgs, "./cmd/xgo/upgrade/...")
		}
	case testKind_xgoTest:
		testArgs = []string{"run", "-tags", "dev", "./cmd/xgo", "test"}
		testArgs = append(testArgs, globalFlags...)
		testArgs = append(testArgs, args...)
		if len(tests) > 0 {
			testArgs = append(testArgs, tests...)
		} else {
			testArgs = append(testArgs, "./test/xgo_test/...")
		}
	case testKind_runtimeTest:
		testArgs = []string{"run", "-tags", "dev", "./cmd/xgo", "test", "--project-dir", "runtime"}
		testArgs = append(testArgs, args...)
		if len(tests) > 0 {
			testArgs = append(testArgs, tests...)
		} else {
			testArgs = append(testArgs,
				"./core/...",
				"./functab/...",
				"./trace/...",
				"./trap/...",
				"./mock/...",
				"./tls/...",
			)
		}
	case testKind_runtimeSubTest:
		if !usePlainGo {
			testArgs = []string{"run", "-tags", "dev", "./cmd/xgo", "test", "--project-dir", "runtime/test"}
		} else {
			testArgs = []string{"test"}
		}

		testArgs = append(testArgs, globalFlags...)
		testArgs = append(testArgs, args...)
		if len(tests) > 0 {
			testArgs = append(testArgs, tests...)
		} else {
			for _, runtimeTest := range runtimeSubTests {
				testArgs = append(testArgs, "./"+runtimeTest+"/...")
			}
		}
	case testKind_xgoAny:
		if !usePlainGo {
			testArgs = []string{"run", "-tags", "dev", "./cmd/xgo", "test"}
		} else {
			testArgs = []string{"test"}
		}

		testArgs = append(testArgs, globalFlags...)
		testArgs = append(testArgs, args...)
		testArgs = append(testArgs, tests...)

	}

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
	fmt.Printf("test Args: %v\n", testArgs)

	execCmd := exec.Command(filepath.Join(goroot, "bin", "go"), testArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Dir = dir

	execCmd.Env = os.Environ()
	execCmd.Env = append(execCmd.Env, "GOROOT="+goroot)
	execCmd.Env = append(execCmd.Env, "PATH="+filepath.Join(goroot, "bin")+string(filepath.ListSeparator)+os.Getenv("PATH"))

	return execCmd.Run()
}
