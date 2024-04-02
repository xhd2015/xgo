package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	cmd_exec "github.com/xhd2015/xgo/support/cmd"
)

// invoking examples:
//
//	go1.21.7/pkg/tool/darwin_amd64/compile -o /var/.../_pkg_.a -trimpath /var/...=> -p fmt -std -complete -buildid b_xx -goversion go1.21.7 -c=4 -nolocalimports -importcfg /var/.../importcfg -pack src/A.go src/B.go
//	go1.21.7/pkg/tool/darwin_amd64/link -V=full
func main() {
	// os.Arg[0] = exec_tool
	// os.Arg[1] = compile or others...
	args := os.Args[1:]

	opts, err := parseOptions(args, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "exec_tool: %v\n", err)
		os.Exit(1)
	}

	logCompileEnable := opts.logCompile
	if logCompileEnable {
		err = initLog()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		// close the file at exit
		defer compileLogFile.Close()
	}

	// log compile args
	if logCompileEnable {
		logCompile("exec_tool %s\n", strings.Join(args, " "))
	}

	toolArgs := opts.remainArgs

	var cmd string

	if len(toolArgs) > 0 {
		cmd = toolArgs[0]
		toolArgs = toolArgs[1:]
	}

	if cmd == "" {
		fmt.Fprintf(os.Stderr, "exec_tool missing command\n")
		os.Exit(1)
	}

	// on Windows, cmd ends with .exe
	baseName := filepath.Base(cmd)
	isCompile := baseName == "compile"
	if !isCompile && runtime.GOOS == "windows" {
		isCompile = baseName == "compile.exe"
	}
	if !isCompile {
		// invoke the process as is
		runCommandExit(cmd, toolArgs)
		return
	}
	err = handleCompile(cmd, opts, toolArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handleCompile(cmd string, opts *options, args []string) error {
	if hasFlag(args, "-V") {
		runCommandExit(cmd, args)
		return nil
	}
	debugWithDlv := opts.debugWithDlv
	// pkg path: the argument after the -p
	pkgPath := findArgAfterFlag(args, "-p")
	if pkgPath == "" {
		return fmt.Errorf("compile missing -p package")
	}
	debugPkg := opts.debug
	var isDebug bool
	if debugPkg != "" {
		if debugPkg == "all" || debugPkg == pkgPath {
			isDebug = true
		}
	}

	xgoCompilerEnableEnv, ok := os.LookupEnv(XGO_COMPILER_ENABLE)
	if !ok {
		if opts.enable {
			xgoCompilerEnableEnv = "true"
		}
	}

	compilerBin := os.Getenv(XGO_COMPILER_BIN)
	if compilerBin == "" {
		compilerBin = cmd
	}

	var withDebugHint string
	if isDebug {
		withDebugHint = " with debug"
	}

	logCompileEnable := opts.logCompile
	if logCompileEnable {
		logCompile("compile %s%s\n", pkgPath, withDebugHint)
	}
	if isDebug {
		// TODO: add env
		if logCompileEnable || debugWithDlv {
			dlvArgs := []string{"--api-version=2",
				"--listen=localhost:2345",
				"--check-go-version=false",
				"--log=true",
				// "--accept-multiclient", // exits when client exits
				"--headless", "exec",
				compilerBin,
				"--",
			}
			envs := getDebugEnv(xgoCompilerEnableEnv)
			// dlvArgs = append(dlvArgs, compilerBin)
			dlvArgs = append(dlvArgs, args...)
			var strPrint []string
			envList := make([]string, 0, len(envs))
			for k, v := range envs {
				envKV := fmt.Sprintf("%s=%s", k, v)
				if logCompileEnable {
					strPrint = append(strPrint, fmt.Sprintf("  export %s", envKV))
				}
				envList = append(envList, envKV)
			}
			if logCompileEnable {
				strPrint = append(strPrint, fmt.Sprintf("  dlv %s", strings.Join(dlvArgs, " ")))
				logCompile("to debug with dlv:\n%s\n", strings.Join(strPrint, "\n"))
			}
			if debugWithDlv {
				logCompile("connect to dlv using CLI: dlv connect localhost:2345\n")
				logCompile("connect to dlv remote using vscode: %s\n", vscodeRemoteDebug)
				var redir io.Writer
				if compileLogFile != nil {
					redir = compileLogFile
				} else {
					redir = io.Discard
				}
				err := cmd_exec.New().Env(envList).Stderr(redir).Stdout(redir).Run("dlv", dlvArgs...)
				if err != nil {
					return fmt.Errorf("dlv: %w", err)
				}
				for {
					time.Sleep(60 * time.Hour)
				}
			}
		}
		debugCmd := getVscodeDebugCmd(compilerBin, xgoCompilerEnableEnv, args)
		debugCmdJSON, err := json.MarshalIndent(debugCmd, "", "    ")
		if err != nil {
			return err
		}
		if logCompileEnable {
			logCompile("%s\n", string(debugCmdJSON))
		}
		vscodeFile := os.Getenv(XGO_DEBUG_VSCODE)

		var nowait bool
		var nowrite bool
		qIdx := strings.LastIndex(vscodeFile, "?")
		if qIdx >= 0 {
			flags := vscodeFile[qIdx+1:]
			vscodeFile = vscodeFile[:qIdx]
			if flags == "nowait" {
				nowait = true
			} else if flags == "nowrite" {
				nowrite = true
			}
		}
		if !nowrite && vscodeFile != "" {
			err = addVscodeDebug(vscodeFile, debugCmd)
			if err != nil {
				return err
			}
		}
		if !nowait {
			// wait the vscode debugger to finish the command
			// sleep forever
			for {
				time.Sleep(60 * time.Hour)
			}
		}
	}
	// COMPILER_ALLOW_SYNTAX_REWRITE=${COMPILER_ALLOW_SYNTAX_REWRITE:-true} COMPILER_ALLOW_IR_REWRITE=${COMPILER_ALLOW_IR_REWRITE:-true} "$shdir/compile" ${@:2}
	runCommandExitFilter(compilerBin, args, func(cmd *exec.Cmd) {
		cmd.Env = append(cmd.Env,
			"XGO_COMPILER_ENABLE="+xgoCompilerEnableEnv,
			// "XGO_COMPILER_ENABLE=false",
			"COMPILER_ALLOW_SYNTAX_REWRITE=true",
			"COMPILER_ALLOW_IR_REWRITE=true",
		)
	})
	return nil
}

func hasFlag(args []string, flag string) bool {
	flagEq := flag + "="
	for _, arg := range args {
		if arg == flag || strings.HasPrefix(arg, flagEq) {
			return true
		}
	}
	return false
}

func findArgAfterFlag(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

func runCommandExit(name string, args []string) {
	runCommandExitFilter(name, args, nil)
}
func runCommandExitFilter(name string, args []string, f func(cmd *exec.Cmd)) {
	err := runCommand(name, args, true, f)
	if err != nil {
		panic(fmt.Errorf("unexpected err: %w", err))
	}
}
func runCommand(name string, args []string, exit bool, f func(cmd *exec.Cmd)) error {
	// invoke the process as is
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if f != nil {
		f(cmd)
	}

	err := cmd.Run()
	if err != nil {
		if exit {
			os.Exit(cmd.ProcessState.ExitCode())
		}
		return err
	}
	return nil
}
