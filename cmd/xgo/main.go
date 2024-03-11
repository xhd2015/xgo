package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// usage:
//   xgo build main.go
//   xgo build .
//   xgo run main
//
// low level flags:
//   -disable-trap          disable trap
//   -disable-runtime-link  disable runtime link

func main() {
	args := os.Args[1:]

	var cmd string
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	if cmd == "" {
		fmt.Fprintf(os.Stderr, "requires cmd\n")
		os.Exit(1)
	}
	if cmd != "build" {
		fmt.Fprintf(os.Stderr, "only support build cmd now, given: %s\n", cmd)
		os.Exit(1)
	}

	err := handleBuild(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handleBuild(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	buildArgs := opts.remainArgs
	flagA := opts.flagA
	projectDir := opts.projectDir
	output := opts.output
	verbose := opts.verbose
	optXgoSrc := opts.xgoSrc
	debug := opts.debug
	vscode := opts.vscode
	withGoroot := opts.withGoroot

	goroot := withGoroot
	if goroot == "" {
		goroot = runtime.GOROOT()
	}
	if goroot == "" {
		return fmt.Errorf("requires GOROOT or --with-goroot")
	}

	if vscode == "" {
		f, err := os.Stat(".vscode")
		if err == nil && f.IsDir() {
			vscode = ".vscode"
		}
	}

	if vscode != "" {
		var err error
		vscode, err = filepath.Abs(vscode)
		if err != nil {
			return err
		}
	}

	// build the exec tool
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get config under home directory: %v", err)
	}
	// check if we are using expected go version
	goVersion, err := getGoVersion(goroot)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(goVersion, "go version go1.20.1") {
		return fmt.Errorf("expect go1.20.1")
	}
	xgoDir := filepath.Join(homeDir, ".xgo")
	buildCacheDir := filepath.Join(xgoDir, "build-cache")
	srcDir := filepath.Join(xgoDir, "src")
	binDir := filepath.Join(xgoDir, "bin")
	logDir := filepath.Join(xgoDir, "log")
	instrumentDir := filepath.Join(xgoDir, "go-instrument")

	instrumentCompilerBin := filepath.Join(binDir, "compile")
	execToolBin := filepath.Join(binDir, "exec_tool")
	compileLog := filepath.Join(logDir, "compile.log")

	err = assertDir(srcDir)
	if err != nil {
		return fmt.Errorf("checking ~/.xgo/src: %w", err)
	}

	err = os.MkdirAll(binDir, 0755)
	if err != nil {
		return fmt.Errorf("create ~/.xgo/bin: %w", err)
	}
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		return fmt.Errorf("create ~/.xgo/log: %w", err)
	}

	if verbose {
		go tailLog(compileLog)
	}
	realXgoSrc := srcDir
	if optXgoSrc != "" {
		realXgoSrc = optXgoSrc
	}

	instrumentGoroot := filepath.Join(instrumentDir, "go1.20.1")
	// patch go runtime and compiler
	err = patchGoSrc(instrumentGoroot, realXgoSrc)
	if err != nil {
		return err
	}

	// build the instrumented compiler
	err = buildCompiler(instrumentGoroot, instrumentCompilerBin)
	if err != nil {
		return err
	}
	// build exec tool
	buildExecToolCmd := exec.Command("go", "build", "-o", execToolBin, "./exec_tool")
	buildExecToolCmd.Dir = filepath.Join(realXgoSrc, "cmd")
	buildExecToolCmd.Stdout = os.Stdout
	buildExecToolCmd.Stderr = os.Stderr
	err = buildExecToolCmd.Run()
	if err != nil {
		return err
	}
	execToolCmd := []string{execToolBin, "--enable"}
	if debug != "" {
		execToolCmd = append(execToolCmd, "--debug="+debug)
	}
	// always add trailing '--' to mark exec tool flags end
	execToolCmd = append(execToolCmd, "--")

	// GOCACHE="$shdir/build-cache" PATH=$goroot/bin:$PATH GOROOT=$goroot DEBUG_PKG=$debug go build -toolexec="$shdir/exce_tool $cmd" "${build_flags[@]}" "$@"
	buildCmdArgs := []string{"build", "-toolexec=" + strings.Join(execToolCmd, " ")}
	if flagA {
		buildCmdArgs = append(buildCmdArgs, "-a")
	}
	if output != "" {
		realOut := output
		if projectDir != "" {
			// make absolute
			absOutput, err := filepath.Abs(output)
			if err != nil {
				return fmt.Errorf("make output absolute: %w", err)
			}
			realOut = absOutput
		}
		buildCmdArgs = append(buildCmdArgs, "-o", realOut)
	}
	buildCmdArgs = append(buildCmdArgs, buildArgs...)
	buildCmd := exec.Command(filepath.Join(instrumentGoroot, "bin", "go"), buildCmdArgs...)
	buildCmd.Env = append(os.Environ(), "GOCACHE="+buildCacheDir)
	buildCmd.Env = patchEnvWithGoroot(buildCmd.Env, instrumentGoroot)
	if vscode != "" {
		buildCmd.Env = append(buildCmd.Env, "XGO_DEBUG_VSCODE="+vscode)
	}
	buildCmd.Env = append(buildCmd.Env, "XGO_COMPILER_BIN="+instrumentCompilerBin)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if projectDir != "" {
		buildCmd.Dir = projectDir
	}
	err = buildCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func buildCompiler(goroot string, output string) error {
	cmd := exec.Command(filepath.Join(goroot, "bin", "go"), "build", "-gcflags=all=-N -l", "-o", output, "./")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = patchEnvWithGoroot(os.Environ(), goroot)
	cmd.Dir = filepath.Join(goroot, "src", "cmd", "compile")
	return cmd.Run()
}

func patchEnvWithGoroot(env []string, goroot string) []string {
	return append(env,
		"GOROOT="+goroot,
		fmt.Sprintf("PATH=%s%c%s", filepath.Join(goroot, "bin"), filepath.ListSeparator, os.Getenv("PATH")),
	)
}

func getGoVersion(goroot string) (string, error) {
	out, err := exec.Command(filepath.Join(goroot, "bin", "go"), "version").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func assertDir(dir string) error {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return err
		// return fmt.Errorf("stat ~/.xgo/src: %v", err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("not a dir")
	}
	return nil
}

// if [[ $verbose = true ]];then
//
//	    tail -fn1 "$shdir/compile.log" &
//	    trap "kill -9 $!" EXIT
//	fi
func tailLog(logFile string) {
	file, err := os.OpenFile(logFile, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open compile log: %v\n", err)
		return
	}
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "seek tail compile log: %v\n", err)
		return
	}
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			fmt.Fprintf(os.Stderr, "tail compile log: %v\n", err)
			return
		}
	}
}
