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

	goroot := runtime.GOROOT()
	if goroot == "" {
		fmt.Fprintf(os.Stderr, "requires GOROOT\n")
		os.Exit(1)
	}
	err := handleBuild(goroot, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handleBuild(goroot string, args []string) error {
	var buildArgs []string
	var flagA bool
	var projectDir string
	var output string
	var verbose bool
	nArg := len(args)
	for i := 0; i < nArg; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			buildArgs = append(buildArgs, arg)
			continue
		}
		if arg == "--" {
			buildArgs = append(buildArgs, args[i+1:]...)
			break
		}
		if arg == "-a" {
			flagA = true
			continue
		}
		if arg == "-v" {
			verbose = true
			continue
		}
		if arg == "--project-dir" {
			if i+1 >= nArg {
				return fmt.Errorf("--project-dir requires argument")
			}
			projectDir = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(arg, "--project-dir=") {
			projectDir = strings.TrimPrefix(arg, "--project-dir=")
			continue
		}
		if arg == "-o" || arg == "--output" {
			if i+1 >= nArg {
				return fmt.Errorf("%s requires argument", arg)
			}
			output = args[i+1]
			i++
			continue
		}
		return fmt.Errorf("unrecognized flag:%s", arg)
	}
	// build the exec tool
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get config under home directory: %v", err)
	}

	xgoDir := filepath.Join(homeDir, ".xgo")
	buildCacheDir := filepath.Join(xgoDir, "build-cache")
	srcDir := filepath.Join(xgoDir, "src")
	binDir := filepath.Join(xgoDir, "bin")
	logDir := filepath.Join(xgoDir, "log")

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

	buildExecToolCmd := exec.Command("go", "build", "-o", execToolBin, "./exec_tool")
	buildExecToolCmd.Dir = filepath.Join(srcDir, "cmd")
	buildExecToolCmd.Stdout = os.Stdout
	buildExecToolCmd.Stderr = os.Stderr
	err = buildExecToolCmd.Run()
	if err != nil {
		return err
	}
	if verbose {
		go tailLog(compileLog)
	}
	// GOCACHE="$shdir/build-cache" PATH=$goroot/bin:$PATH GOROOT=$goroot DEBUG_PKG=$debug go build -toolexec="$shdir/exce_tool $cmd" "${build_flags[@]}" "$@"
	buildCmdArgs := []string{"build", "-toolexec=" + execToolBin}
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
	buildCmd := exec.Command("go", buildCmdArgs...)
	buildCmd.Env = append(os.Environ(), "GOCACHE="+buildCacheDir)
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
