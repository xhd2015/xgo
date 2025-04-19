package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/script/xgo.helper/instrument"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/goinfo"
)

const help = `
xgo.helper

Commands:
  setup-vscode GOROOT                setup vscode settings.json and launch.json
  debug-go GOROOT <cmd>              debug go build,go run, go test
  run-go GOROOT <cmd>                run go build,go run, go test
  debug-compile GOROOT <pkg> <cmd>   debug compile <pkg> <cmd>

Options:
  -C dir     passed to go

Examples:
  $ xgo.helper setup-vscode GOROOT
  $ xgo.helper debug-go GOROOT -C $X/xgo/runtime/test/patch/real_world/kusia_ipc test -v ./
  $ xgo.helper debug-compile GOROOT some.pkg -C $X/xgo/runtime/test/patch/real_world/kusia_ipc test -v ./
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires cmd, try 'help'")
	}
	cmd := args[0]
	if cmd == "help" || cmd == "--help" || cmd == "-h" {
		fmt.Println(strings.TrimSpace(help))
		return nil
	}
	args = args[1:]
	switch cmd {
	case "debug-go", "run-go":
		if len(args) == 0 || args[0] == "" {
			return fmt.Errorf("requires GOROOT")
		}

		goroot := args[0]
		err := instrumentPkgLoad(goroot)
		if err != nil {
			return err
		}
		err = instrumentModLoad(goroot)
		if err != nil {
			return err
		}
		return debugGo(cmd == "run-go", goroot, nil, args[1:])
	case "debug-compile":
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("requires GOROOT")
		}
		goroot := args[0]
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("requires pkg")
		}
		if strings.HasPrefix(args[1], "-") {
			return fmt.Errorf("invalid pkg: %s", args[1])
		}
		goVersionStr, err := goinfo.GetGoVersionOutput(filepath.Join(goroot, "bin", "go"))
		if err != nil {
			return err
		}
		goVersion, err := goinfo.ParseGoVersion(goVersionStr)
		if err != nil {
			return err
		}
		err = instrument.InstrumentGc(goroot, goVersion)
		if err != nil {
			return err
		}
		return debugGo(true, goroot, []string{"XGO_HELPER_DEBUG_PKG=" + args[1]}, args[2:])
	case "setup-vscode":
		if len(args) == 0 || args[0] == "" {
			return fmt.Errorf("requires GOROOT")
		}
		return setupVscode(args[0], args[1:])
	}
	return fmt.Errorf("unknown cmd: %s", cmd)
}

func debugGo(runOnly bool, goroot string, env []string, args []string) error {
	if goroot == "" {
		return fmt.Errorf("requires GOROOT")
	}
	var cmdArgs []string
	var flagC string

	var runCmd string
	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			if arg != "build" && arg != "test" && arg != "run" {
				return fmt.Errorf("bad cmd: %s", arg)
			}
			runCmd = arg
			cmdArgs = append(cmdArgs, args[i+1:]...)
			break
		}
		if arg == "-C" {
			if i+1 >= n {
				return fmt.Errorf("requires -C")
			}
			flagC = args[i+1]
			i++
			continue
		} else if strings.HasPrefix(arg, "-C=") {
			flagC = arg[len("-C="):]
			continue
		}
		return fmt.Errorf("unrecognized: %s", arg)
	}
	goroot, err := checkedAbsGoroot(goroot)
	if err != nil {
		return err
	}

	if runCmd == "" {
		return fmt.Errorf("requires cmd: build,test or run")
	}

	goExe, err := build.BuildGoDebugBinary(goroot)
	if err != nil {
		return err
	}
	_, err = build.BuildGoToolCompileDebugBinary(goroot)
	if err != nil {
		return err
	}
	return runOrDlvExec(goroot, goExe, runOnly, runCmd, flagC, env, cmdArgs)
}

func setupVscode(goroot string, args []string) error {
	if goroot == "" {
		return fmt.Errorf("requires GOROOT")
	}
	// n := len(args)
	// for i := 0; i < n; i++ {
	// 	arg := args[i]
	// 	if arg == "--goroot" {
	// 		if i+1 >= n {
	// 			return fmt.Errorf("requires goroot")
	// 		}
	// 		goroot = args[i+1]
	// 		i++
	// 		continue
	// 	} else if strings.HasPrefix(arg, "--goroot=") {
	// 		goroot = arg[len("--goroot="):]
	// 		continue
	// 	}
	// 	return fmt.Errorf("unrecognized: %s", arg)
	// }
	goroot, err := checkedAbsGoroot(goroot)
	if err != nil {
		return err
	}
	vscodeDir := filepath.Join(goroot, ".vscode")
	err = os.MkdirAll(vscodeDir, 0755)
	if err != nil {
		return err
	}
	err = writeNonExistingFile(filepath.Join(vscodeDir, "settings.json"), []byte(fmt.Sprintf(`{
		"go.goroot": %q,
	}`, goroot)))
	if err != nil {
		return err
	}

	err = writeNonExistingFile(filepath.Join(vscodeDir, "launch.json"), []byte(`{
    "configurations": [
        {
            "name": "Debug dlv localhost:2345(go)",
            "type": "go",
            "debugAdapter": "dlv-dap",
            "request": "attach",
            "mode": "remote",
            "port": 2345,
            "host": "127.0.0.1",
            "cwd": "./"
        },
		{
            "name": "Debug dlv localhost:2346(compile)",
            "type": "go",
            "debugAdapter": "dlv-dap",
            "request": "attach",
            "mode": "remote",
            "port": 2346,
            "host": "127.0.0.1",
            "cwd": "./"
        }
    ]
}`))
	if err != nil {
		return err
	}
	return nil
}

func writeNonExistingFile(path string, content []byte) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return os.WriteFile(path, content, 0755)
	}
	fmt.Fprintf(os.Stderr, "file %s already exists, skip\n", path)
	return nil
}
func runOrDlvExec(goroot string, goExe string, runOnly bool, runCmd string, flagC string, env []string, args []string) error {
	// dlv exec --listen=:2345 --api-version=2 --check-go-version=false --headless -- go.debug test --project-dir runtime/test -v ./patch
	var cmdExe string
	var cmdArgs []string
	if !runOnly {
		cmdExe = "dlv"
		cmdArgs = []string{"exec", "--listen=:2345", "--api-version=2", "--check-go-version=false", "--headless", "--", goExe}
	} else {
		cmdExe = goExe
	}
	if flagC != "" {
		cmdArgs = append(cmdArgs, "-C", flagC)
	}
	cmdArgs = append(cmdArgs, runCmd)
	cmdArgs = append(cmdArgs, args...)
	runEnv := make([]string, 0, len(env)+2)
	runEnv = append(runEnv,
		"GOROOT="+goroot,
		"GO_BYPASS_XGO=true",
	)
	runEnv = append(runEnv, env...)
	return cmd.Debug().Env(runEnv).Run(cmdExe, cmdArgs...)
}

func checkedAbsGoroot(goroot string) (string, error) {
	if goroot == "" {
		goroot = runtime.GOROOT()
		if goroot == "" {
			return "", fmt.Errorf("cannot find GOROOT")
		}
	}
	srcRuntime := filepath.Join(goroot, "src", "runtime")
	_, err := os.Stat(srcRuntime)
	if err != nil {
		return "", fmt.Errorf("GOROOT is not valid: %w", err)
	}
	absGoroot, err := filepath.Abs(goroot)
	if err != nil {
		return "", err
	}
	return absGoroot, nil
}
