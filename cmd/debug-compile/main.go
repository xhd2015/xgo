package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/xgo/cmd/xgo/exec_tool"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/debug"
	"github.com/xhd2015/xgo/support/netutil"
)

// usage:
//  go run -tags dev ./cmd/xgo --debug-compile ./src   --> will generate a file called debug-compile.json
//  go run -tags dev ./cmd/xgo build --build-compile   --> will build the compiler with -gcflags and print it's path
//  go run ./cmd/debug-compile --debug-with-dlv        --> will read debug-compile.json, and start a debug server listen on localhost:2345

func main() {
	args := os.Args[1:]
	err := run(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var compilerBinary string
	n := len(args)
	var projectDir string

	data, readErr := ioutil.ReadFile("debug-compile.json")
	if readErr != nil {
		if !errors.Is(readErr, os.ErrNotExist) {
			return readErr
		}
	}
	var debugCompiler *exec_tool.DebugCompile
	if len(data) > 0 {
		err := json.Unmarshal(data, &debugCompiler)
		if err != nil {
			return fmt.Errorf("parse debug-compile.json: %w", err)
		}
	}

	var extraFlags []string
	var extraEnvs []string
	var debugWithDlv bool
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--project-dir" {
			projectDir = args[i+1]
			i++
			continue
		}
		if arg == "--env" {
			extraEnvs = append(extraEnvs, args[i+1])
			i++
			continue
		}
		if arg == "--compiler" {
			compilerBinary = args[i+1]
			i++
			continue
		}
		if arg == "-cpuprofile" {
			extraFlags = append(extraFlags, arg, args[i+1])
			i++
			continue
		}
		if arg == "-blockprofile" {
			extraFlags = append(extraFlags, arg, args[i+1])
			i++
			continue
		}
		if arg == "-memprofile" || arg == "-memprofilerate" {
			extraFlags = append(extraFlags, arg, args[i+1])
			i++
			continue
		}
		if arg == "-N" || arg == "-l" {
			extraFlags = append(extraFlags, arg)
			continue
		}
		if arg == "-c" {
			extraFlags = append(extraFlags, arg, args[i+1])
			continue
		}
		if arg == "--debug-with-dlv" {
			debugWithDlv = true
			continue
		}
	}
	if compilerBinary == "" {
		compilerBinary = debugCompiler.Compiler
	}

	runArgs := append(debugCompiler.Flags, extraFlags...)
	runArgs = append(runArgs, debugCompiler.Files...)

	runEnvs := append(debugCompiler.Env, extraEnvs...)

	if debugWithDlv {
		return dlvExecListen(projectDir, runEnvs, compilerBinary, runArgs)
	}

	start := time.Now()
	defer func() {
		fmt.Printf("cost: %v\n", time.Since(start))
	}()
	return cmd.Dir(projectDir).Env(runEnvs).Run(compilerBinary, runArgs...)
}

// /Users/xhd2015/.xgo/go-instrument-dev/go1.21.7_Us_xh_in_go_096be049/compile

func dlvExecListen(dir string, env []string, compilerBinary string, args []string) error {
	var vscodeExtra []string
	n := len(env)
	for i := n - 1; i >= 0; i-- {
		e := env[i]
		if !strings.HasPrefix(e, "GOROOT=") {
			continue
		}
		goroot := strings.TrimPrefix(e, "GOROOT=")
		if goroot != "" {
			vscodeExtra = append(vscodeExtra,
				fmt.Sprintf("    NOTE: VSCode will map source files to workspace's goroot,"),
				fmt.Sprintf("      To fix this, update .vscode/settings.json, set go.goroot to:"),
				fmt.Sprintf("       %s", goroot),
				fmt.Sprintf("      And set a breakpoint at:"),
				fmt.Sprintf("       %s/src/cmd/compile/main.go", goroot),
			)
		}
		break
	}
	return netutil.ServePort(2345, true, 500*time.Millisecond, func(port int) {
		prompt := debug.FormatDlvPromptOptions(port, &debug.FormatDlvOptions{
			VscodeExtra: vscodeExtra,
		})
		fmt.Println(prompt)
	}, func(port int) error {
		// dlv exec --api-version=2 --listen=localhost:2345 --accept-multiclient --headless ./debug.bin
		dlvArgs := []string{
			"--api-version=2",
			fmt.Sprintf("--listen=localhost:%d", port),
			"--check-go-version=false",
			"--log=true",
			// "--accept-multiclient", // exits when client exits
			"--headless", "exec",
			compilerBinary,
			"--",
		}
		// dlvArgs = append(dlvArgs, compilerBin)
		dlvArgs = append(dlvArgs, args...)
		err := cmd.Dir(dir).Env(env).Run("dlv", dlvArgs...)
		if err != nil {
			return fmt.Errorf("dlv: %w", err)
		}
		return nil
	})
}
