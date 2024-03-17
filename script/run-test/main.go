package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]
	var excludes []string
	var includes []string
	n := len(args)
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
		err := runTest(filepath.Join(goRelease, goroot))
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

func runTest(goroot string) error {
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}

	execCmd := exec.Command(filepath.Join(goroot, "bin", "go"), "test", "./test")
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	execCmd.Env = os.Environ()
	execCmd.Env = append(execCmd.Env, "GOROOT="+goroot)
	execCmd.Env = append(execCmd.Env, "PATH="+filepath.Join(goroot, "bin")+string(filepath.ListSeparator)+os.Getenv("PATH"))

	return execCmd.Run()
}
