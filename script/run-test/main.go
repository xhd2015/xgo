package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

// usage:
//
//	go run ./script/run-test/ --include go1.19.13
//	go run ./script/run-test/ --include go1.19.13 -run TestHelloWorld -v
//	go run ./script/run-test/ --include go1.17.13 --include go1.18.10 --include go1.19.13 --include go1.20.14 --include go1.21.8 --include go1.22.1 -count=1
func main() {
	args := os.Args[1:]
	var excludes []string
	var includes []string

	n := len(args)
	var remainArgs []string

	var resetInstrument bool
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
		if arg == "--" {
			if i+1 < n {
				remainArgs = append(remainArgs, args[i+1:]...)
			}
			break
		}
	}
	goRelease := "go-release"
	goroots, err := listGoroots(goRelease)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if len(includes) > 0 {
		i := 0
		for _, goroot := range goroots {
			if listContains(includes, goroot) {
				goroots[i] = goroot
				i++
			}
		}
		goroots = goroots[:i]
	} else if len(excludes) > 0 {
		i := 0
		for _, goroot := range goroots {
			if !listContains(excludes, goroot) {
				goroots[i] = goroot
				i++
			}
		}
		goroots = goroots[:i]
	}

	if len(goroots) == 0 {
		fmt.Fprintf(os.Stderr, "no go select\n")
		os.Exit(1)
	}
	for _, goroot := range goroots {
		fmt.Fprintf(os.Stdout, "TEST %s\n", goroot)
		if resetInstrument {
			err := cmd.Run("go", "run", "./cmd/xgo", "--reset-instrument", "--with-goroot", goroot, "--build-compiler")
			if err != nil {
				if extErr, ok := err.(*exec.ExitError); ok {
					fmt.Fprintf(os.Stderr, "%s\n", extErr.Stderr)
				} else {
					fmt.Fprintf(os.Stderr, "%v", err)
				}
				os.Exit(1)
			}
		}
		err := runTest(filepath.Join(goRelease, goroot), remainArgs)
		if err != nil {
			fmt.Fprintf(os.Stdout, "FAIL %s: %v\n", goroot, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "PASS %s\n", goroot)
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

func runTest(goroot string, args []string) error {
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}

	testArgs := []string{"test"}
	if len(args) > 0 {
		testArgs = append(testArgs, args...)
	}
	testArgs = append(testArgs, "./test")

	execCmd := exec.Command(filepath.Join(goroot, "bin", "go"), testArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	execCmd.Env = os.Environ()
	execCmd.Env = append(execCmd.Env, "GOROOT="+goroot)
	execCmd.Env = append(execCmd.Env, "PATH="+filepath.Join(goroot, "bin")+string(filepath.ListSeparator)+os.Getenv("PATH"))

	return execCmd.Run()
}
