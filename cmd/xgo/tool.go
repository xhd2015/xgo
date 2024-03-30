package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

func handleTool(tool string, args []string) error {
	tools := []string{
		tool,
	}
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(tool, ".exe") {
			tools = append(tools, tool+".exe")
		} else {
			tools = append(tools, strings.TrimSuffix(tool, ".exe"))
		}
	}
	var lastErr error
	n := len(tools)
	for i := n - 1; i >= 0; i-- {
		var invoked bool
		invoked, lastErr = tryHandleTool(tools[i], args)
		if invoked {
			return lastErr
		}
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func tryHandleTool(tool string, args []string) (invoked bool, err error) {
	baseName := filepath.Base(tool)
	if baseName != tool {
		return false, fmt.Errorf("unknown tool: %s", tool)
	}
	curName := filepath.Base(os.Args[0])
	if baseName == curName {
		// cannot invoke itself
		return false, fmt.Errorf("unknown tool: %s", tool)
	}
	dirName := filepath.Dir(os.Args[0])
	toolExec := filepath.Join(dirName, tool)

	var retryHome bool
	stat, statErr := os.Stat(toolExec)
	if statErr != nil {
		if !errors.Is(statErr, os.ErrNotExist) {
			return false, fmt.Errorf("unknown tool: %s", tool)
		}
		retryHome = true
	} else if stat.IsDir() {
		retryHome = true
	}
	if retryHome {
		// try ~/.xgo/bin/tool
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return false, fmt.Errorf("unknown tool: %s", tool)
		}
		toolExec = filepath.Join(home, ".xgo", "bin", tool)
		stat, statErr := os.Stat(toolExec)
		if statErr != nil || stat.IsDir() {
			return false, fmt.Errorf("unknown tool: %s", tool)
		}
	}

	return true, cmd.Run(toolExec, args...)
}
