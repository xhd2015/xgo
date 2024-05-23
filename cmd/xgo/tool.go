package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/coverage"
	test_explorer "github.com/xhd2015/xgo/cmd/xgo/test-explorer"
	"github.com/xhd2015/xgo/cmd/xgo/trace"
	"github.com/xhd2015/xgo/support/cmd"
)

const toolHelp = `
Xgo ships with a toolset that improves testing experience.

Usage:
    xgo tool <command> [arguments]

The commands are:
    trace          stack trace visualization
    test-explorer  test explorer
    coverage       incremental coverage tool
    list           list all tools
    help           show help

Examples:
    xgo tool trace TestSomething.json     visualize a generated trace
    xgo tool test-explorer                open test explorer UI
    xgo tool coverage serve cover.out     visualize incrementa coverage of cover.out

See https://github.com/xhd2015/xgo for documentation.

`

const toolList = `
Available tools:
  trace            visualize a generated trace
  test-explorer    open test explorer UI
  coverage         visualize incrementa coverage 
`

func handleTool(args []string) error {
	// tool string,

	// if len(args) == 0 {
	// 	fmt.Fprintf(os.Stderr, "xgo tool: requires tool to run\n")
	// 	os.Exit(1)
	// }

	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(toolHelp, "\n"))
		return nil
	}
	tool := args[0]
	args = args[1:]
	if tool == "list" {
		fmt.Print(strings.TrimPrefix(toolList, "\n"))
		return nil
	}
	if tool == "trace" {
		trace.Main(args)
		return nil
	}
	if tool == "coverage" {
		coverage.Main(args)
		return nil
	}
	if tool == "test-explorer" {
		return test_explorer.Main(args, &test_explorer.Options{
			DefaultGoCommand: "xgo",
		})
	}
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
