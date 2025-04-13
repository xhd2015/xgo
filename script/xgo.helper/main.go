package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/osinfo"
)

const help = `
xgo.helper

Commands:
  debug-go GOROOT build       debug go build,go run, go test
  create-vscode GOROOT        create vscode settings.json and launch.json

Options:
  -C dir     passed to go

Examples:
  $ xgo.helper create-vscode GOROOT
  $ xgo.helper debug-go GOROOT -C $X/xgo/runtime/test/patch/real_world/kusia_ipc test -v ./
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
	case "debug-go":
		if len(args) == 0 || args[0] == "" {
			return fmt.Errorf("requires GOROOT")
		}
		return debugGo(args[0], args[1:])
	case "create-vscode":
		if len(args) == 0 || args[0] == "" {
			return fmt.Errorf("requires GOROOT")
		}
		return createVscode(args[0], args[1:])
	}
	return fmt.Errorf("unknown cmd: %s", cmd)
}

func debugGo(goroot string, args []string) error {
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

	goExe, err := buildGo(goroot)
	if err != nil {
		return err
	}
	return dlvExecGo(goroot, goExe, runCmd, flagC, cmdArgs)
}

func createVscode(goroot string, args []string) error {
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
            "name": "Debug dlv localhost:2345",
            "type": "go",
            "debugAdapter": "dlv-dap",
            "request": "attach",
            "mode": "remote",
            "port": 2345,
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

func buildGo(goroot string) (string, error) {
	absGoroot, err := filepath.Abs(goroot)
	if err != nil {
		return "", err
	}
	src := filepath.Join(absGoroot, "src")
	bin := filepath.Join(absGoroot, "bin")
	output := filepath.Join(bin, "go.debug"+osinfo.EXE_SUFFIX)
	err = cmd.Debug().Dir(src).Env([]string{"GOROOT=" + absGoroot, "GO_BYPASS_XGO=true"}).Run(filepath.Join(bin, "go"+osinfo.EXE_SUFFIX), "build", "-gcflags=all=-N -l", "-o", output, "./cmd/go")
	if err != nil {
		return "", fmt.Errorf("build go: %w", err)
	}
	return output, nil
}

func dlvExecGo(goroot string, goExe string, runCmd string, flagC string, args []string) error {
	// dlv exec --listen=:2345 --api-version=2 --check-go-version=false --headless -- go.debug test --project-dir runtime/test -v ./patch
	dlvArgs := []string{"exec", "--listen=:2345", "--api-version=2", "--check-go-version=false", "--headless", "--", goExe}
	if flagC != "" {
		dlvArgs = append(dlvArgs, "-C", flagC)
	}
	dlvArgs = append(dlvArgs, runCmd)
	dlvArgs = append(dlvArgs, args...)
	return cmd.Debug().Dir(goroot).Env([]string{"GOROOT=" + goroot, "GO_BYPASS_XGO=true"}).Run("dlv", dlvArgs...)
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
