package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var exeSuffix = getExeSuffix()

func getExeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func main() {
	// if invoked with go, os.Args[0] = "go"
	realGo, err := GetRealGoPath(os.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	args := os.Args[1:]
	var subCmd string
	if len(args) > 0 {
		subCmd = args[0]
	}
	var cmdErr error
	if os.Getenv("XGO_SHADOW_BYPASS") != "true" && (subCmd == "build" || subCmd == "run" || subCmd == "test") {
		cmdErr = runCmd("xgo", args, []string{"XGO_REAL_GO_BINARY=" + realGo})
	} else {
		cmdErr = runCmd(realGo, args, nil)
	}

	if cmdErr != nil {
		var exitErr *exec.ExitError
		if errors.As(cmdErr, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "%v\n", cmdErr)
		os.Exit(1)
	}
}

func runCmd(cmd string, args []string, env []string) error {
	exeCmd := exec.Command(cmd, args...)
	exeCmd.Stdout = os.Stdout
	exeCmd.Stderr = os.Stderr
	exeCmd.Env = os.Environ()
	exeCmd.Env = append(exeCmd.Env, env...)
	return exeCmd.Run()
}

func GetRealGoPath(curGo string) (string, error) {
	var absGo string
	var err error
	// if curGo == "go"
	if curGo == "go" || curGo == "go"+exeSuffix {
		absGo, err = exec.LookPath(curGo)
		if err != nil {
			return "", err
		}
	} else {
		if !filepath.IsAbs(curGo) {
			return "", fmt.Errorf("xgo shadow not called with go or absolute path: %s", curGo)
		}
		absGo = curGo
	}
	return GetRealGoPathExcludeDir(filepath.Dir(absGo))
}

func GetRealGoPathExcludeDir(dir string) (string, error) {
	callback, err := exclude(dir)
	if err != nil {
		return "", err
	}
	defer callback()

	return exec.LookPath("go")
}
