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
//	go run ./script/run-test/ --include go1.19.13 -count=1
//	go run ./script/run-test/ --include go1.19.13 -run TestHelloWorld -v
//	go run ./script/run-test/ --include go1.17.13 --include go1.18.10 --include go1.19.13 --include go1.20.14 --include go1.21.8 --include go1.22.1 -count=1
//  go run ./script/run-test/ -cover -coverpkg github.com/xhd2015/xgo/runtime/... -coverprofile covers/cover.out --include go1.21.8

// runtime test:
//     go run ./script/run-test/ --include go1.17.13 --xgo-runtime-test-only -run TestFuncList -v ./test/func_list

// xgo default test:
//    go run ./script/run-test/ --include go1.18.10 --xgo-default-test-only -run TestAtomicGenericPtr -v ./test

// xgo test
//    go run ./script/run-test/ --include go1.18.10 --xgo-test-only -run TestFuncNames -v ./test/xgo_test/func_names

// run specific test for all go versions
//     go run ./script/run-test/ --include go1.17.13 --include go1.18.10 --include go1.19.13 --include go1.20.14 --include go1.21.8 --include go1.22.1 -count=1 --xgo-default-test-only -run TestFuncNameSliceShouldWorlWithDebug -v

// TODO: remove duplicate test between xgo test and runtime test
var runtimeTests = []string{
	"trace_panic_peek",
	"func_list",
	"trap_inspect_func",
	"trap",
	"mock_func",
	"mock_method",
	"mock_by_name",
	"mock_closure",
	"mock_stdlib",
	"mock_generic",
	"trap_args",
	"patch",
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
	var xgoDefaultTestOnly bool

	var debug bool
	var cover bool
	var coverpkgs []string
	var coverprofile string
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
			xgoDefaultTestOnly = false
			continue
		}
		if arg == "--xgo-runtime-test-only" {
			xgoTestOnly = false
			xgoRuntimeTestOnly = true
			xgoDefaultTestOnly = false
			continue
		}
		if arg == "--xgo-default-test-only" {
			xgoTestOnly = false
			xgoRuntimeTestOnly = false
			xgoDefaultTestOnly = true
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
			if debug {
				fmt.Printf("reseting instrument\n")
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
		if xgoTestOnly {
			runRuntimeTestFlag = false
			runXgoTestFlag = true
			runDefault = false
		} else if xgoRuntimeTestOnly {
			runRuntimeTestFlag = true
			runXgoTestFlag = false
			runDefault = false
		} else if xgoDefaultTestOnly {
			runRuntimeTestFlag = false
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
			if debug && (variant == "xgo" || variant == "runtime") {
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
	return doRunTest(goroot, testKind_default, args, tests)
}

func runXgoTest(goroot string, args []string, tests []string) error {
	return doRunTest(goroot, testKind_xgoTest, args, tests)
}

func runRuntimeTest(goroot string, args []string, tests []string) error {
	return doRunTest(goroot, testKind_runtimeTest, args, tests)
}

type testKind string

const (
	testKind_default     testKind = ""
	testKind_xgoTest     testKind = "xgo-test"
	testKind_runtimeTest testKind = "runtime-test"
)

func doRunTest(goroot string, kind testKind, args []string, tests []string) error {
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}
	var testArgs []string
	switch kind {
	case testKind_default:
		testArgs = []string{"test"}
		testArgs = append(testArgs, args...)
		if len(tests) > 0 {
			testArgs = append(testArgs, tests...)
		} else {
			testArgs = append(testArgs, "./test")
			testArgs = append(testArgs, "./support/...")
			testArgs = append(testArgs, "./cmd/...")
		}
	case testKind_xgoTest:
		testArgs = []string{"run", "./cmd/xgo", "test"}
		testArgs = append(testArgs, args...)
		if len(tests) > 0 {
			testArgs = append(testArgs, tests...)
		} else {
			testArgs = append(testArgs, "./test/xgo_test/...")
		}
	case testKind_runtimeTest:
		testArgs = []string{"run", "./cmd/xgo", "test", "--project-dir", "runtime"}
		testArgs = append(testArgs, args...)
		if len(tests) > 0 {
			testArgs = append(testArgs, tests...)
		} else {
			for _, runtimeTest := range runtimeTests {
				testArgs = append(testArgs, "./test/"+runtimeTest)
			}
			testArgs = append(testArgs,
				"./core/...",
				"./functab/...",
				"./trace/...",
				"./trap/...",
				"./mock/...",
			)
		}
	}

	// debug
	// fmt.Printf("test Args: %v\n", testArgs)

	execCmd := exec.Command(filepath.Join(goroot, "bin", "go"), testArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	execCmd.Env = os.Environ()
	execCmd.Env = append(execCmd.Env, "GOROOT="+goroot)
	execCmd.Env = append(execCmd.Env, "PATH="+filepath.Join(goroot, "bin")+string(filepath.ListSeparator)+os.Getenv("PATH"))

	return execCmd.Run()
}
