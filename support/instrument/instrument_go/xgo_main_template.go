// This file will be added to instrumented $GOROOT/src/cmd/go/xgo_main.go
// the 'go:build ignore' line will be removed when copied

//go:build ignore

package main

import (
	"errors"
	"os"
	"os/exec"
)

// pass to xgo unless we cannot find xgo
func xgoPrecheck(cmd string, osArgs []string) bool {
	if cmd != "build" && cmd != "test" && cmd != "run" {
		return false
	}
	if os.Getenv("GO_BYPASS_XGO") == "true" {
		return false
	}
	// pass to xgo unless we cannot find xgo
	args := osArgs[1:]
	xgoCmd := exec.Command("xgo", args...)
	xgoCmd.Stdout = os.Stdout
	xgoCmd.Stderr = os.Stderr
	xgoCmd.Env = append(os.Environ(), "GO_BYPASS_XGO=true")
	runErr := xgoCmd.Run()
	if runErr == nil {
		return true
	}
	if !errors.Is(runErr, exec.ErrNotFound) {
		return false
	}
	exitCode := 1
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		exitCode = exitErr.ExitCode()
	}
	os.Exit(exitCode)
	return true
}
