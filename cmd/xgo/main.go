package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/osinfo"

	"github.com/xhd2015/xgo/cmd/xgo/exec_tool"
	"github.com/xhd2015/xgo/cmd/xgo/pathsum"
	"github.com/xhd2015/xgo/support/goinfo"
)

// usage:
//   xgo build main.go
//   xgo build .
//   xgo run main
//   xgo exec go version
//
// low level flags:
//   -disable-trap          disable trap
//   -disable-runtime-link  disable runtime link

var closeDebug func()

func main() {
	args := os.Args[1:]

	var cmd string
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	if cmd == "" {
		fmt.Fprintf(os.Stderr, "xgo: requires command\nRun 'xgo help' for usage.\n")
		os.Exit(1)
	}
	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return
	}
	if cmd == "version" {
		fmt.Println(VERSION)
		return
	}
	if cmd == "revision" {
		fmt.Println(getRevision())
		return
	}
	if cmd == "upgrade" {
		err := handleUpgrade(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}
	if cmd == "exec_tool" {
		exec_tool.Main(args)
		return
	}
	if cmd == "tool" {
		err := handleTool(args)
		if err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				os.Exit(err.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}
	if cmd != "build" && cmd != "run" && cmd != "test" && cmd != "exec" {
		fmt.Fprintf(os.Stderr, "xgo %s: unknown command\nRun 'xgo help' for usage.\n", cmd)
		os.Exit(1)
	}
	defer func() {
		if closeDebug != nil {
			defer closeDebug()
		}
	}()

	err := handleBuild(cmd, args)
	if err != nil {
		var errMsg string
		var exitCode int = 1
		if e, ok := err.(*exec.ExitError); ok {
			errMsg = string(e.Stderr)
			if errMsg == "" {
				errMsg = err.Error()
			}
			exitCode = e.ExitCode()
		} else {
			errMsg = err.Error()
		}
		logDebug("finished with error: %s", errMsg)
		fmt.Fprintf(os.Stderr, "%v\n", errMsg)
		os.Exit(exitCode)
	}
	logDebug("finished successfully")
}

func handleBuild(cmd string, args []string) error {
	cmdBuild := cmd == "build"
	cmdTest := cmd == "test"
	cmdExec := cmd == "exec"
	opts, err := parseOptions(cmd, args)
	if err != nil {
		return err
	}
	remainArgs := opts.remainArgs
	testArgs := opts.testArgs
	flagA := opts.flagA
	projectDir := opts.projectDir
	output := opts.output
	flagV := opts.flagV
	flagX := opts.flagX
	flagC := opts.flagC
	logCompile := opts.logCompile
	logDebugOption := opts.logDebug
	optXgoSrc := opts.xgoSrc
	noBuildOutput := opts.noBuildOutput
	noInstrument := opts.noInstrument
	resetInstrument := opts.resetInstrument
	noSetup := opts.noSetup
	debugWithDlv := opts.debugWithDlv
	xgoHome := opts.xgoHome
	syncXgoOnly := opts.syncXgoOnly
	setupDev := opts.setupDev
	buildCompiler := opts.buildCompiler
	syncWithLink := opts.syncWithLink
	debug := opts.debug
	vscode := opts.vscode
	mod := opts.mod
	gcflags := opts.gcflags
	overlay := opts.overlay
	modfile := opts.modfile
	withGoroot := opts.withGoroot
	dumpIR := opts.dumpIR
	dumpAST := opts.dumpAST
	stackTrace := opts.stackTrace
	trapStdlib := opts.trapStdlib

	if cmdExec && len(remainArgs) == 0 {
		return fmt.Errorf("exec requires command")
	}

	closeDebug, err = setupDebugLog(logDebugOption)
	if err != nil {
		return err
	}
	if logDebugOption != nil {
		logStartup()
	}

	goroot, err := checkGoroot(projectDir, withGoroot)
	if err != nil {
		return err
	}
	// make the goroot abs
	goroot, err = filepath.Abs(goroot)
	if err != nil {
		return err
	}
	logDebug("effective GOROOT: %s", goroot)

	// create a tmp dir for communication with exec_tool
	tmpDir, err := os.MkdirTemp("", "xgo-"+cmd)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	var vscodeDebugFile string
	var vscodeDebugFileSuffix string
	if !noInstrument && debug != "" {
		var err error
		vscodeDebugFile, vscodeDebugFileSuffix, err = getVscodeDebugFile(tmpDir, vscode)
		if err != nil {
			return err
		}
	}

	var tmpIRFile string
	var tmpASTFile string
	if !noInstrument {
		if dumpIR != "" {
			tmpIRFile = filepath.Join(tmpDir, "dump-ir")
		}
		if dumpAST != "" {
			tmpASTFile = filepath.Join(tmpDir, "dump-ast")
		}
	}

	// build the exec tool
	goVersion, err := checkGoVersion(goroot, noInstrument)
	if err != nil {
		return err
	}

	goVersionName := fmt.Sprintf("go%d.%d.%d", goVersion.Major, goVersion.Minor, goVersion.Patch)
	logDebug("go version: %s", goVersionName)
	mappedGorootName, err := pathsum.PathSum(goVersionName+"_", goroot)
	if err != nil {
		return err
	}

	// abs xgo dir
	xgoDir, err := getXgoHome(xgoHome)
	if err != nil {
		return err
	}
	binDir := filepath.Join(xgoDir, "bin")
	logDir := filepath.Join(xgoDir, "log")
	instrumentSuffix := ""
	if isDevelopment {
		instrumentSuffix = "-dev"
	}
	instrumentDir := filepath.Join(xgoDir, "go-instrument"+instrumentSuffix, mappedGorootName)
	logDebug("instrument dir: %s", instrumentDir)

	exeSuffix := osinfo.EXE_SUFFIX
	// NOTE: on Windows, go build -o xxx will always yield xxx.exe
	compileLog := filepath.Join(logDir, "compile.log")
	compilerBin := filepath.Join(instrumentDir, "compile"+exeSuffix)
	compilerBuildID := filepath.Join(instrumentDir, "compile.buildid.txt")
	instrumentGoroot := filepath.Join(instrumentDir, goVersionName)
	packageDataDir := filepath.Join(instrumentDir, "pkgdata")

	// gcflags can cause the build cache to invalidate
	// so separate them with normal one
	buildCacheSuffix := ""
	if trapStdlib {
		buildCacheSuffix += "-trapstd"
	}
	if len(gcflags) > 0 {
		buildCacheSuffix += "-gcflags"
	}
	buildCacheDir := filepath.Join(instrumentDir, "build-cache"+buildCacheSuffix)
	revisionFile := filepath.Join(instrumentDir, "xgo-revision.txt")
	fullSyncRecord := filepath.Join(instrumentDir, "full-sync-record.txt")

	var realXgoSrc string
	if isDevelopment {
		realXgoSrc = optXgoSrc
		if realXgoSrc == "" {
			_, statErr := os.Stat(filepath.Join("cmd", "xgo", "main.go"))
			if statErr == nil {
				realXgoSrc, _ = os.Getwd()
			}
		}
		if realXgoSrc == "" {
			return fmt.Errorf("requires --xgo-src")
		}
	}

	setupDone := make(chan struct{})
	go func() {
		d := 1 * time.Second
		if isDevelopment {
			d = 5 * time.Second
		}
		select {
		case <-time.After(d):
			fmt.Fprintf(os.Stderr, "xgo is taking a while to setup, please wait...\n")
		case <-setupDone:
		}
	}()

	var revisionChanged bool
	if !noInstrument && !noSetup {
		err = ensureDirs(binDir, logDir, instrumentDir, packageDataDir)
		if err != nil {
			return err
		}
		revision := getRevision()
		if isDevelopment {
			revision = revision + " DEV"
		}
		var err error
		revisionChanged, err = checkRevisionChanged(revisionFile, revision)
		if err != nil {
			return err
		}

		if resetInstrument {
			logDebug("revision changed, reset %s", instrumentDir)
			err := os.RemoveAll(instrumentDir)
			if err != nil {
				return err
			}
		}
		if isDevelopment || resetInstrument || revisionChanged {
			logDebug("sync goroot %s -> %s", goroot, instrumentGoroot)
			err = syncGoroot(goroot, instrumentGoroot, fullSyncRecord)
			if err != nil {
				return err
			}
			// patch go runtime and compiler
			logDebug("patch compiler at: %s", instrumentGoroot)
			err = patchRuntimeAndCompiler(goroot, instrumentGoroot, realXgoSrc, goVersion, syncWithLink || setupDev || buildCompiler, revisionChanged)
			if err != nil {
				return err
			}
		}

		if resetInstrument || revisionChanged {
			logDebug("revision %s write to %s", revision, revisionFile)
			err := os.WriteFile(revisionFile, []byte(revision), 0755)
			if err != nil {
				return err
			}
		}
	}

	if isDevelopment && setupDev {
		logDebug("setup dev dir: %s", instrumentGoroot)
		err := setupDevDir(instrumentGoroot, setupDev)
		if err != nil {
			return err
		}
		if setupDev {
			return nil
		}
	}
	if syncXgoOnly {
		return nil
	}

	var compilerChanged bool
	var toolExecFlag string
	if !noInstrument {
		logDebug("build instrument tools: %s", instrumentGoroot)
		xgoBin := os.Args[0]
		compilerChanged, toolExecFlag, err = buildInstrumentTool(instrumentGoroot, realXgoSrc, compilerBin, compilerBuildID, "", xgoBin, debug, logCompile, noSetup, debugWithDlv)
		if err != nil {
			return err
		}
		logDebug("compiler changed: %v", compilerChanged)
		logDebug("tool exec flags: %v", toolExecFlag)
	}
	close(setupDone)

	if buildCompiler {
		return nil
	}

	// before invoking exec_tool, tail follow its log
	if logCompile {
		go tailLog(compileLog)
	}

	execCmdEnv := os.Environ()
	var execCmd *exec.Cmd
	if !cmdExec {
		if modfile != "" {
			// make modfile absolute
			modfile, err = filepath.Abs(modfile)
			if err != nil {
				return nil
			}
		}
		instrumentGo := filepath.Join(instrumentGoroot, "bin", "go"+osinfo.EXE_SUFFIX)
		subPaths, mainModule, err := goinfo.ResolveMainModule(projectDir, remainArgs)
		if err != nil {
			if !errors.Is(err, goinfo.ErrGoModNotFound) && !errors.Is(err, goinfo.ErrGoModDoesNotHaveModule) {
				return err
			}
		}
		if stackTrace == "on" && overlay == "" {
			// check if xgo/runtime ready
			impResult, impRuntimeErr := importRuntimeDep(cmdTest, instrumentGoroot, instrumentGo, goVersion, modfile, realXgoSrc, projectDir, subPaths, mainModule, mod, remainArgs)
			if impRuntimeErr != nil {
				// can be silently ignored
				fmt.Fprintf(os.Stderr, "WARNING: --strace requires: import _ %q\n   failed to auto import %s: %v\n", RUNTIME_TRACE_PKG, RUNTIME_TRACE_PKG, impRuntimeErr)
			} else if impResult != nil {
				overlay = impResult.overlayFile
				if impResult.mod != "" {
					mod = impResult.mod
				}
				if impResult.modfile != "" {
					modfile = impResult.modfile
				}
			}
		}
		logDebug("resolved main module: %s", mainModule)
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_MAIN_MODULE+"="+mainModule)
		// GOCACHE="$shdir/build-cache" PATH=$goroot/bin:$PATH GOROOT=$goroot DEBUG_PKG=$debug go build -toolexec="$shdir/exce_tool $cmd" "${build_flags[@]}" "$@"
		buildCmdArgs := []string{cmd}
		if toolExecFlag != "" {
			buildCmdArgs = append(buildCmdArgs, toolExecFlag)
		}
		if flagA || compilerChanged || revisionChanged {
			buildCmdArgs = append(buildCmdArgs, "-a")
		}
		if flagV {
			buildCmdArgs = append(buildCmdArgs, "-v")
		}
		if flagX {
			buildCmdArgs = append(buildCmdArgs, "-x")
		}
		if flagC {
			buildCmdArgs = append(buildCmdArgs, "-c")
		}
		if overlay != "" {
			buildCmdArgs = append(buildCmdArgs, "-overlay", overlay)
		}
		if mod != "" {
			buildCmdArgs = append(buildCmdArgs, "-mod="+mod)
		}
		if modfile != "" {
			buildCmdArgs = append(buildCmdArgs, "-modfile", modfile)
		}

		for _, f := range gcflags {
			buildCmdArgs = append(buildCmdArgs, "-gcflags="+f)
		}
		if cmdBuild || (cmdTest && flagC) {
			// output
			if noBuildOutput {
				discardOut := filepath.Join(tmpDir, "discard.out")
				buildCmdArgs = append(buildCmdArgs, "-o", discardOut)
			} else if output != "" {
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
		}
		buildCmdArgs = append(buildCmdArgs, remainArgs...)
		if cmdTest && len(testArgs) > 0 {
			buildCmdArgs = append(buildCmdArgs, testArgs...)
		}
		logDebug("command: %s %v", instrumentGo, buildCmdArgs)
		execCmd = exec.Command(instrumentGo, buildCmdArgs...)
	} else {
		logDebug("command: %v", remainArgs)
		execCmd = exec.Command(remainArgs[0], remainArgs[1:]...)
	}
	execCmd.Env = execCmdEnv
	execCmd.Env, err = patchEnvWithGoroot(execCmd.Env, instrumentGoroot)
	if err != nil {
		return err
	}
	if !noInstrument {
		execCmd.Env = append(execCmd.Env, "GOCACHE="+buildCacheDir)
		execCmd.Env = append(execCmd.Env, "XGO_COMPILER_BIN="+compilerBin)
		execCmd.Env = append(execCmd.Env, exec_tool.XGO_COMPILE_PKG_DATA_DIR+"="+packageDataDir)
		// xgo versions
		execCmd.Env = append(execCmd.Env, "XGO_TOOLCHAIN_VERSION="+VERSION)
		execCmd.Env = append(execCmd.Env, "XGO_TOOLCHAIN_REVISION="+REVISION)
		execCmd.Env = append(execCmd.Env, "XGO_TOOLCHAIN_VERSION_NUMBER="+strconv.FormatInt(NUMBER, 10))

		// IR
		execCmd.Env = append(execCmd.Env, "XGO_DEBUG_DUMP_IR="+dumpIR)
		execCmd.Env = append(execCmd.Env, "XGO_DEBUG_DUMP_IR_FILE="+tmpIRFile)

		// AST
		execCmd.Env = append(execCmd.Env, "XGO_DEBUG_DUMP_AST="+dumpAST)
		execCmd.Env = append(execCmd.Env, "XGO_DEBUG_DUMP_AST_FILE="+tmpASTFile)

		// vscode debug
		var xgoDebugVscode string
		if vscodeDebugFile != "" {
			xgoDebugVscode = vscodeDebugFile + vscodeDebugFileSuffix
		}
		execCmd.Env = append(execCmd.Env, "XGO_DEBUG_VSCODE="+xgoDebugVscode)

		// stack trace
		if stackTrace != "" {
			execCmd.Env = append(execCmd.Env, "XGO_STACK_TRACE="+stackTrace)
		}

		// trap stdlib
		var trapStdlibEnv string
		if trapStdlib {
			trapStdlibEnv = "true"
		}
		execCmd.Env = append(execCmd.Env, "XGO_STD_LIB_TRAP_DEFAULT_ALLOW="+trapStdlibEnv)
	}
	logDebug("command env: %v", execCmd.Env)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	if projectDir != "" {
		execCmd.Dir = projectDir
	}
	logDebug("command dir: %v", execCmd.Dir)
	logDebug("command executable path: %v", execCmd.Path)
	err = execCmd.Run()
	if err != nil {
		return err
	}

	// if dump IR is not nil, output to stdout
	if tmpIRFile != "" {
		err := copyToStdout(tmpIRFile)
		if err != nil {
			// ir file not exists
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("--dump-ir not effective, use -a to trigger recompile or check if you package path matches")
			}
			return err
		}
	}
	if tmpASTFile != "" {
		err := copyToStdout(tmpASTFile)
		if err != nil {
			// ir file not exists
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("--dump-ast not effective, use -a to trigger recompile or check if you package path matches")
			}
			return err
		}
	}
	if (vscode == "stdout" || strings.HasPrefix(vscode, "stdout?")) && vscodeDebugFile != "" {
		err := copyToStdout(vscodeDebugFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("vscode debug not effective, use -a to trigger recompile")
			}
			return err
		}
	}
	return nil
}

func checkGoVersion(goroot string, noInstrument bool) (*goinfo.GoVersion, error) {
	// check if we are using expected go version
	goVersionStr, err := goinfo.GetGoVersionOutput(filepath.Join(goroot, "bin", "go"))
	if err != nil {
		return nil, err
	}
	goVersion, err := goinfo.ParseGoVersion(goVersionStr)
	if err != nil {
		return nil, err
	}
	if !noInstrument {
		minor := goVersion.Minor
		if goVersion.Major != 1 || (minor < 17 || minor > 22) {
			return nil, fmt.Errorf("only supports go1.17.0 ~ go1.22.1, current: %s", goVersionStr)
		}
	}
	return goVersion, nil
}

// getXgoHome returns absolute path to xgo homoe
func getXgoHome(xgoHome string) (string, error) {
	if xgoHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get config under home directory: %v", err)
		}
		return filepath.Join(homeDir, ".xgo"), nil
	}

	absHome, err := filepath.Abs(xgoHome)
	if err != nil {
		return "", fmt.Errorf("make abs --xgo-home %s: %w", xgoHome, err)
	}
	absHomeDir := filepath.Dir(absHome)
	if absHomeDir == "." || absHomeDir == "" || absHomeDir == "/" || absHomeDir == absHome {
		return "", fmt.Errorf("invalid --xgo-home: %s", xgoHome)
	}
	_, err = os.Stat(absHomeDir)
	if err != nil {
		return "", fmt.Errorf("not found --xgo-home %s: %w", xgoHome, err)
	}
	return absHome, nil
}

func getGoEnvRoot(dir string) (string, error) {
	goroot, err := cmd.Dir(dir).Output("go", "env", "GOROOT")
	if err != nil {
		return "", err
	}
	// remove special characters from output
	goroot = strings.ReplaceAll(goroot, "\n", "")
	goroot = strings.ReplaceAll(goroot, "\r", "")
	return goroot, nil
}

func checkGoroot(dir string, goroot string) (string, error) {
	if goroot == "" {
		// use env first because dir will affect the actual
		// go version used
		// because with new go toolchain mechanism, even with:
		//    which go -> /Users/xhd2015/installed/go1.21.7/bin/go
		// the actual go still points to go1.22.0
		//    go version ->
		// go tool chain: https://go.dev/doc/toolchain
		// GOTOOLCHAIN=auto
		envGoroot, envErr := getGoEnvRoot(dir)
		if envErr == nil && envGoroot != "" {
			return envGoroot, nil
		}
		goroot = runtime.GOROOT()
		if goroot != "" {
			return goroot, nil
		}

		if envErr != nil {
			var errMsg string
			if e, ok := envErr.(*exec.ExitError); ok {
				errMsg = string(e.Stderr)
			} else {
				errMsg = envErr.Error()
			}
			return "", fmt.Errorf("requires GOROOT or --with-goroot: go env GOROOT: %v", errMsg)
		}
		return "", fmt.Errorf("requires GOROOT or --with-goroot")
	}
	_, err := os.Stat(goroot)
	if err == nil {
		return goroot, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if filepath.Base(goroot) != goroot {
		return "", err
	}
	// check if inside go-release
	releaseGo := filepath.Join("go-release", goroot)
	releaseStat, statErr := os.Stat(releaseGo)
	if statErr == nil && releaseStat.IsDir() {
		return releaseGo, nil
	}
	if strings.HasPrefix(goroot, "go") {
		fmt.Fprintf(os.Stderr, "WARNING %s does not exist, download it with:\n          go run ./script/download-go %s\n", goroot, goroot)
	}
	return "", err
}

func ensureDirs(binDir string, logDir string, instrumentDir string, packageDataDir string) error {
	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		return fmt.Errorf("create ~/.xgo/bin: %w", err)
	}
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		return fmt.Errorf("create ~/.xgo/log: %w", err)
	}
	err = os.MkdirAll(instrumentDir, 0755)
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(instrumentDir), err)
	}
	err = os.MkdirAll(packageDataDir, 0755)
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(packageDataDir), err)
	}
	return nil
}

func patchEnvWithGoroot(env []string, goroot string) ([]string, error) {
	goroot, err := filepath.Abs(goroot)
	if err != nil {
		return nil, err
	}
	return append(env,
		"GOROOT="+goroot,
		fmt.Sprintf("PATH=%s%c%s", filepath.Join(goroot, "bin"), filepath.ListSeparator, os.Getenv("PATH")),
	), nil
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
