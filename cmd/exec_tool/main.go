package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	err := initLog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	// os.Arg[0] = exec_tool
	// os.Arg[1] = compile or others...
	args := os.Args[1:]

	// log compile args
	logCompile("%s\n", strings.Join(args, " "))

	var cmd string

	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	if cmd == "" {
		fmt.Fprintf(os.Stderr, "exec_tool missing command\n")
		os.Exit(1)
	}

	baseName := filepath.Base(cmd)
	if baseName != "compile" {
		// invoke the process as is
		runCommandExit(cmd, args)
		return
	}
	err = handleCompile(cmd, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handleCompile(cmd string, args []string) error {
	// pkg=$(echo "$@"|grep -oE -- "-p .*?( |$)"|cut -c 4-|tr -d ' ') # take package
	// log_compile compile $pkg
	// runCommand(compileCmd, args)

	runCommandExit(cmd, args)

	return nil
}

func runCommandExit(name string, args []string) {
	err := runCommand(name, args, true)
	if err != nil {
		panic(fmt.Errorf("unexpected err: %w", err))
	}
}
func runCommand(name string, args []string, exit bool) error {
	// invoke the process as is
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		if exit {
			os.Exit(cmd.ProcessState.ExitCode())
		}
		return err
	}
	return nil
}
