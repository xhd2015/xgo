package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	cmd_support "github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/instrument/overlay"

	debug_support "github.com/xhd2015/xgo/support/debug"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/netutil"
	"github.com/xhd2015/xgo/support/osinfo"

	"github.com/xhd2015/xgo/cmd/xgo/exec_tool"
	"github.com/xhd2015/xgo/cmd/xgo/pathsum"
	test_explorer "github.com/xhd2015/xgo/cmd/xgo/test-explorer"
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

// since xgo V1_0.1.0, xgo does not instrument the compiler anymore

const V1_0_0 = false

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
		if len(args) > 0 && args[0] == "--core" {
			fmt.Println(getCoreRevision())
		} else {
			fmt.Println(getRevision())
		}
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
	if cmd == "test-explorer" || cmd == "explorer" || cmd == "e" {
		err := test_explorer.Main(args, &test_explorer.Options{
			DefaultGoCommand: "xgo",
		})
		consumeErrAndExit(err)
		return
	}
	if cmd == "tool" {
		err := handleTool(args)
		consumeErrAndExit(err)
		return
	}
	if cmd == "shadow" {
		err := handleShadow()
		consumeErrAndExit(err)
		return
	}
	if cmd != "build" && cmd != "run" && cmd != "test" && cmd != "exec" {
		fmt.Fprintf(os.Stderr, "xgo %s: unknown command\nRun 'xgo help' for usage.\n", cmd)
		os.Exit(1)
	}
	defer func() {
		if closeDebug != nil {
			// always close the debug log regardless of panic
			defer closeDebug()
		}
	}()

	err := handleBuild(cmd, args)
	if err != nil {
		code, stderr := parseErr(err)
		logDebug("finished with error: %s", stderr)
		if len(stderr) > 0 {
			os.Stderr.Write(stderr)
			os.Stderr.WriteString("\n")
		}
		os.Exit(code)
		return
	}
	logDebug("finished successfully")
}

func handleBuild(cmd string, args []string) error {
	cmdRun := cmd == "run"
	cmdBuild := cmd == "build"
	cmdTest := cmd == "test"
	cmdExec := cmd == "exec"
	opts, err := parseOptions(cmd, args)
	if err != nil {
		return err
	}
	remainArgs := opts.remainArgs
	testArgs := opts.testArgs
	buildFlags := opts.buildFlags
	progFlags := opts.progFlags
	flagA := opts.flagA
	projectDir := opts.projectDir
	output := opts.output
	flagV := opts.flagV
	flagX := opts.flagX
	flagC := opts.flagC
	flagRun := opts.flagRun
	logCompile := opts.logCompile
	logDebugOption := opts.logDebug
	debugCompile := opts.debugCompile
	debug := opts.debug
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
	debugTarget := opts.debugTarget
	vscode := opts.vscode
	mod := opts.mod
	gcflags := opts.gcflags
	overlayFile := opts.overlay
	modfile := opts.modfile
	withGoroot := opts.withGoroot
	dumpIR := opts.dumpIR
	dumpAST := opts.dumpAST
	stackTrace := opts.stackTrace
	stackTraceDir := opts.stackTraceDir
	straceSnapshotMainModuleDefault := opts.straceSnapshotMainModuleDefault
	trapStdlib := opts.trapStdlib
	trapPkgs := opts.trap
	noLineDirective := opts.noLineDirective

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
	tmpRoot, err := getTmpDir()
	if err != nil {
		return err
	}
	// the xgoTmpDir can be thought as a shortly-stable cache dir
	//  - it won't gets deleted in a short period
	//  - the system will reclaim it if out of storage
	// it will hold:
	//  - instrumentation
	//  - build cache
	xgoTmpDir := filepath.Join(tmpRoot, "xgo")
	logDebug("xgoTmpDir: %s", xgoTmpDir)
	err = os.MkdirAll(xgoTmpDir, 0755)
	if err != nil {
		return err
	}

	// generate at .xgo/gen
	// and add .xgo/gen to .gitignore
	localXgoGenDir, err := getLocalXgoGenDir(projectDir)
	if err != nil {
		return err
	}

	// sessionTmpDir is specifically for this run session
	sessionTmpDir, err := os.MkdirTemp(xgoTmpDir, "session-"+cmd+"-")
	logDebug("sessionTmpDir: %s", sessionTmpDir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(sessionTmpDir)

	optionsFromFile, optionsFromFileContent, err := mergeOptionFiles(sessionTmpDir, opts.optionsFromFile, opts.mockRules)
	if err != nil {
		return err
	}

	var vscodeDebugFile string
	var vscodeDebugFileSuffix string
	if !noInstrument && debugTarget != "" {
		var err error
		vscodeDebugFile, vscodeDebugFileSuffix, err = getVscodeDebugFile(sessionTmpDir, vscode)
		if err != nil {
			return err
		}
	}

	var tmpIRFile string
	var tmpASTFile string
	if !noInstrument {
		if dumpIR != "" {
			tmpIRFile = filepath.Join(sessionTmpDir, "dump-ir")
		}
		if dumpAST != "" {
			tmpASTFile = filepath.Join(sessionTmpDir, "dump-ast")
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
	xgoDir, err := getOrMakeAbsXgoHome(xgoHome)
	if err != nil {
		return err
	}
	binDir := filepath.Join(xgoDir, "bin")
	logDir := filepath.Join(xgoTmpDir, "log")
	instrumentSuffix := ""
	if isDevelopment {
		instrumentSuffix = "-dev"
	}
	instrumentDir := filepath.Join(xgoTmpDir, "go-instrument"+instrumentSuffix, mappedGorootName)
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
	if len(gcflags) > 0 || debug != nil {
		buildCacheSuffix += "-gcflags"
	}
	if optionsFromFile != "" && len(optionsFromFileContent) > 0 {
		h := md5.New()
		h.Write(optionsFromFileContent)
		buildCacheSuffix += "-" + hex.EncodeToString(h.Sum(nil))
	}
	var enableStackTrace bool
	if stackTrace == "on" || stackTrace == "true" {
		enableStackTrace = true
		v := stackTrace
		if v == "true" {
			v = "on"
		} else if v == "false" {
			v = "off"
		}
		buildCacheSuffix += "-strace_" + v
	}
	if stackTraceDir != "" && stackTraceDir != "." && stackTraceDir != "./" {
		// this affects the trap/flags package
		h := md5.New()
		h.Write([]byte(stackTraceDir))
		buildCacheSuffix += "-" + hex.EncodeToString(h.Sum(nil))
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

	var coreRevisionChanged bool
	if !noInstrument && !noSetup {
		err = ensureDirs(binDir, logDir, instrumentDir, packageDataDir)
		if err != nil {
			return err
		}
		coreRevision := getCoreRevision()
		var err error
		coreRevisionChanged, err = checkRevisionChanged(revisionFile, coreRevision)
		if err != nil {
			return err
		}

		if resetInstrument {
			logDebug("reset instrument %s", instrumentDir)
			err := os.RemoveAll(instrumentDir)
			if err != nil {
				return err
			}
		}
		if resetInstrument || flagA {
			logDebug("reset overlay %s", localXgoGenDir)
			err := os.RemoveAll(localXgoGenDir)
			if err != nil {
				return err
			}
		}
		resetOrRevisionChanged := resetInstrument || buildCompiler || coreRevisionChanged
		if isDevelopment || resetOrRevisionChanged {
			logDebug("sync goroot %s -> %s", goroot, instrumentGoroot)
			err = syncGoroot(goroot, instrumentGoroot, fullSyncRecord)
			if err != nil {
				return err
			}
			// patch go runtime and compiler
			logDebug("patch runtime at: %s", instrumentGoroot)
			err = patchRuntimeAndCompiler(goroot, instrumentGoroot, realXgoSrc, goVersion, syncWithLink || setupDev || buildCompiler, resetOrRevisionChanged)
			if err != nil {
				return err
			}
		}

		if resetOrRevisionChanged {
			logDebug("revision %s write to %s", coreRevision, revisionFile)
			err := os.WriteFile(revisionFile, []byte(coreRevision), 0755)
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
	if V1_0_0 {
		if !noInstrument {
			logDebug("build instrument tools: %s", instrumentGoroot)
			xgoBin := os.Args[0]
			compilerChanged, toolExecFlag, err = buildInstrumentTool(instrumentGoroot, realXgoSrc, compilerBin, compilerBuildID, "", xgoBin, debugTarget, logCompile, noSetup, debugWithDlv)
			if err != nil {
				return err
			}
			logDebug("compiler changed: %v", compilerChanged)
			logDebug("tool exec flags: %v", toolExecFlag)
		}
	}
	close(setupDone)

	if V1_0_0 && buildCompiler {
		fmt.Printf("%s\n", compilerBin)
		return nil
	}

	logDebug("trap stdlib: %v", trapStdlib)

	// before invoking exec_tool, tail follow its log
	if logCompile {
		go tailLog(compileLog)
	}
	var debugCompilePkg string
	var debugCompileLogFile string

	var runDebug bool
	var testDebug bool
	var buildDebug bool

	var runFlagsAfterBuild []string
	if debug != nil {
		if cmdRun {
			runDebug = true
		} else if cmdTest {
			if !flagC {
				testDebug = true
			} else {
				buildDebug = true
			}
		} else if cmdBuild {
			buildDebug = true
		} else {
			return fmt.Errorf("invalid --debug flag")
		}
	}

	debugMode := runDebug || testDebug || buildDebug
	var finalBuildOutput string

	execCmdEnv, err := patchEnvWithGoroot(os.Environ(), instrumentGoroot)
	if err != nil {
		return err
	}
	var execCmd *exec.Cmd
	var logCmdExec func()
	if !cmdExec {
		if modfile != "" {
			// make modfile absolute
			modfile, err = filepath.Abs(modfile)
			if err != nil {
				return nil
			}
		}
		instrumentGo := filepath.Join(instrumentGoroot, "bin", "go"+osinfo.EXE_SUFFIX)
		subPaths, mainModule, err := goinfo.ResolveMainModule(projectDir)
		if err != nil {
			if !errors.Is(err, goinfo.ErrGoModNotFound) && !errors.Is(err, goinfo.ErrGoModDoesNotHaveModule) {
				return err
			}
		}
		if debugCompile != nil {
			if *debugCompile != "" {
				debugCompilePkg = *debugCompile
			} else {
				// find the main package we are compile
				pkgs, err := goinfo.ListPackagePaths(projectDir, mod, remainArgs)
				if err != nil {
					return err
				}
				if len(pkgs) == 0 {
					return fmt.Errorf("--debug-compile: no packages")
				}
				if len(pkgs) > 1 {
					return fmt.Errorf("--debug-compile: need 1 package, found: %d", len(pkgs))
				}
				debugCompilePkg = pkgs[0]
			}
			debugCompileLogFile = filepath.Join(sessionTmpDir, "debug-compile.log")
			go tailLog(debugCompileLogFile)
			logDebug("debug compile package: %s", debugCompilePkg)
		}
		overlayFS := overlay.MakeOverlay()
		if overlayFile != "" {
			goOverlay, err := overlay.ReadGoOverlay(overlayFile)
			if err != nil {
				return err
			}
			for k, v := range goOverlay.Replace {
				overlayFS.OverrideFile(k, v)
			}
		}

		projectRoot := getProjectRoot(projectDir, subPaths)
		var xgoRuntimeModuleDir string

		// the most important result of `importRuntimeDepGenOverlay`
		// is the standalone runtime module dir,i.e. `xgoRuntimeModuleDir`
		// except this, there are no other significant new go files
		// intruduced. following steps that loads packages should
		// sticks to old mod and modfile
		modForLoad := mod
		modfileForLoad := modfile
		if enableStackTrace {
			// check if xgo/runtime ready
			impResult, impRuntimeErr := importRuntimeDepGenOverlay(cmdTest, instrumentGoroot, instrumentGo, goVersion, modfile, realXgoSrc, projectDir, projectRoot, localXgoGenDir, mainModule, mod, resetInstrument || flagA, remainArgs)
			if impRuntimeErr != nil {
				// can be silently ignored
				fmt.Fprintf(os.Stderr, "WARNING: --strace requires: import _ %q\n   failed to auto import %s: %v\n", RUNTIME_TRACE_PKG, RUNTIME_TRACE_PKG, impRuntimeErr)
			} else if impResult != nil {
				if impResult.mod != "" {
					mod = impResult.mod
				}
				if impResult.modfile != "" {
					modfile = impResult.modfile
				}
				// override go files, which has auto import _ "github.com/xhd2015/xgo/runtime/trace"
				for src, target := range impResult.fileReplace {
					overlayFS.OverrideFile(src, target)
				}
				// after we have replaced modules, and use -mod=mod
				// go will try to read target modules's go.mod even
				// they don't exist.
				// so here we also put these phantom files in the overlay,
				// they just don't have to exist in vendor
				for mod, target := range impResult.modReplace {
					overlayFS.OverrideFile(mod, target)
				}
				xgoRuntimeModuleDir = impResult.runtimeModuleDir
			}
		}

		var opts FileOptions
		if len(optionsFromFileContent) > 0 {
			err = json.Unmarshal(optionsFromFileContent, &opts)
			if err != nil {
				return err
			}
		}
		// coverage and instrument have conflicts,
		// see https://github.com/xhd2015/xgo/issues/301
		// TODO: enhance this to be extract -coverpkg pkg1,pkg2,...
		var mayHaveCover bool
		// example:
		//    -cover -coverpkg github.com/xhd2015/xgo/... -coverprofile cover.out
		for _, arg := range args {
			if strings.HasPrefix(arg, "-cover") {
				mayHaveCover = true
				break
			}
		}

		var collectTestTrace bool
		var collectTestTraceDir string
		if cmdTest && enableStackTrace {
			collectTestTrace = true
			collectTestTraceDir = stackTraceDir
		}
		err = instrumentUserSpace(projectDir, projectRoot, modForLoad, modfileForLoad, mainModule, xgoRuntimeModuleDir, mayHaveCover, overlayFS, cmdTest, opts.FilterRules, trapPkgs, collectTestTrace, collectTestTraceDir)
		if err != nil {
			return err
		}

		if len(overlayFS) > 0 {
			goOverlay, err := overlayFS.MakeGoOverlay(filepath.Join(localXgoGenDir, "overlay"), !noLineDirective)
			if err != nil {
				return err
			}
			if len(goOverlay.Replace) > 0 {
				overlayFile = filepath.Join(localXgoGenDir, "overlay.json")
				logDebug("write overlay file: %s", overlayFile)
				err = goOverlay.Write(overlayFile)
				if err != nil {
					return err
				}
				if projectDir != "" {
					overlayFile, err = filepath.Abs(overlayFile)
					if err != nil {
						return err
					}
				}
			}
		}
		logDebug("resolved main module: %s", mainModule)
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_MAIN_MODULE+"="+mainModule)
		// GOCACHE="$shdir/build-cache" PATH=$goroot/bin:$PATH GOROOT=$goroot DEBUG_PKG=$debug go build -toolexec="$shdir/exce_tool $cmd" "${build_flags[@]}" "$@"
		var buildCmdArgs []string
		if !runDebug {
			buildCmdArgs = []string{cmd}
		} else {
			buildCmdArgs = []string{"build"}
		}

		if toolExecFlag != "" {
			buildCmdArgs = append(buildCmdArgs, toolExecFlag)
		}
		if flagA || compilerChanged || coreRevisionChanged {
			buildCmdArgs = append(buildCmdArgs, "-a")
		}
		if flagV {
			if !testDebug {
				buildCmdArgs = append(buildCmdArgs, "-v")
			} else {
				runFlagsAfterBuild = append(runFlagsAfterBuild, "-test.v")
			}
		}
		if flagX {
			buildCmdArgs = append(buildCmdArgs, "-x")
		}
		if flagC && !testDebug {
			buildCmdArgs = append(buildCmdArgs, "-c")
		}
		if flagRun != "" {
			if !testDebug {
				buildCmdArgs = append(buildCmdArgs, "-run", flagRun)
			} else {
				runFlagsAfterBuild = append(runFlagsAfterBuild, "-test.run", flagRun)
			}
		}
		if overlayFile != "" {
			buildCmdArgs = append(buildCmdArgs, "-overlay", overlayFile)
		}
		if mod != "" {
			buildCmdArgs = append(buildCmdArgs, "-mod="+mod)
		}
		if modfile != "" {
			buildCmdArgs = append(buildCmdArgs, "-modfile", modfile)
		}
		var hasBuildDebug bool
		for _, f := range gcflags {
			if f == "all=-N -l" || f == "all=-l -N" {
				hasBuildDebug = true
			}
			buildCmdArgs = append(buildCmdArgs, "-gcflags="+f)
		}

		if debugMode {
			if !hasBuildDebug {
				buildCmdArgs = append(buildCmdArgs, "-gcflags=all=-N -l")
			}
			if testDebug {
				buildCmdArgs = append(buildCmdArgs, "-c")
			}
		}

		if cmdBuild || (cmdTest && flagC) || debugMode {
			// output
			if !debugMode && noBuildOutput {
				discardOut := filepath.Join(sessionTmpDir, "discard.out")
				buildCmdArgs = append(buildCmdArgs, "-o", discardOut)
			} else if output != "" {
				finalBuildOutput = output
				if projectDir != "" {
					// make absolute
					absOutput, err := filepath.Abs(output)
					if err != nil {
						return fmt.Errorf("make output absolute: %w", err)
					}
					finalBuildOutput = absOutput
				}
			} else if debugMode {
				finalBuildOutput = filepath.Join(sessionTmpDir, "debug.bin")
			}
			if finalBuildOutput != "" {
				buildCmdArgs = append(buildCmdArgs, "-o", finalBuildOutput)
			}
		}
		if len(buildFlags) > 0 {
			buildCmdArgs = append(buildCmdArgs, buildFlags...)
		}
		if len(progFlags) > 0 {
			runFlagsAfterBuild = append(runFlagsAfterBuild, progFlags...)
		}
		if len(remainArgs) > 0 {
			if !runDebug {
				buildCmdArgs = append(buildCmdArgs, remainArgs...)
			} else {
				buildCmdArgs = append(buildCmdArgs, remainArgs[0])
				runFlagsAfterBuild = append(runFlagsAfterBuild, remainArgs[1:]...)
			}
		}

		if cmdTest && len(testArgs) > 0 {
			if !testDebug {
				buildCmdArgs = append(buildCmdArgs, "-args")
				buildCmdArgs = append(buildCmdArgs, testArgs...)
			} else {
				runFlagsAfterBuild = append(runFlagsAfterBuild, testArgs...)
			}
		}
		logCmdExec = func() {
			logDebug("command: %s %v", instrumentGo, buildCmdArgs)
			if len(runFlagsAfterBuild) > 0 {
				logDebug("prog flags: %v", runFlagsAfterBuild)
			}
		}
		execCmd = exec.Command(instrumentGo, buildCmdArgs...)
	} else {
		logCmdExec = func() {
			logDebug("command: %v", remainArgs)
		}
		execCmd = exec.Command(remainArgs[0], remainArgs[1:]...)
	}
	if !noInstrument {
		execCmdEnv = append(execCmdEnv, "GOCACHE="+buildCacheDir)
	}

	if !noInstrument && V1_0_0 {
		execCmdEnv = append(execCmdEnv, "XGO_COMPILER_BIN="+compilerBin)
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_COMPILE_PKG_DATA_DIR+"="+packageDataDir)
		// xgo versions
		execCmdEnv = append(execCmdEnv, "XGO_TOOLCHAIN_VERSION="+CORE_VERSION)
		execCmdEnv = append(execCmdEnv, "XGO_TOOLCHAIN_REVISION="+CORE_REVISION)
		execCmdEnv = append(execCmdEnv, "XGO_TOOLCHAIN_VERSION_NUMBER="+strconv.FormatInt(CORE_NUMBER, 10))

		// IR
		execCmdEnv = append(execCmdEnv, "XGO_DEBUG_DUMP_IR="+dumpIR)
		execCmdEnv = append(execCmdEnv, "XGO_DEBUG_DUMP_IR_FILE="+tmpIRFile)

		// AST
		execCmdEnv = append(execCmdEnv, "XGO_DEBUG_DUMP_AST="+dumpAST)
		execCmdEnv = append(execCmdEnv, "XGO_DEBUG_DUMP_AST_FILE="+tmpASTFile)

		// vscode debug
		var xgoDebugVscode string
		if vscodeDebugFile != "" {
			xgoDebugVscode = vscodeDebugFile + vscodeDebugFileSuffix
		}
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_DEBUG_VSCODE+"="+xgoDebugVscode)

		// debug compile package
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_DEBUG_COMPILE_PKG+"="+debugCompilePkg)
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_DEBUG_COMPILE_LOG_FILE+"="+debugCompileLogFile)

		// stack trace
		if stackTrace != "" {
			execCmdEnv = append(execCmdEnv, exec_tool.XGO_STACK_TRACE+"="+stackTrace)
		}
		if stackTraceDir != "" {
			execCmdEnv = append(execCmdEnv, exec_tool.XGO_STACK_TRACE_DIR+"="+stackTraceDir)
		}
		if enableStackTrace && straceSnapshotMainModuleDefault != "" {
			execCmdEnv = append(execCmdEnv, exec_tool.XGO_STRACE_SNAPSHOT_MAIN_MODULE_DEFAULT+"="+straceSnapshotMainModuleDefault)
		}

		// trap stdlib
		var trapStdlibEnv string
		if trapStdlib {
			trapStdlibEnv = "true"
		}
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_STD_LIB_TRAP_DEFAULT_ALLOW+"="+trapStdlibEnv)

		// compiler options (make abs)
		var absOptionsFromFile string
		if optionsFromFile != "" {
			absOptionsFromFile, err = filepath.Abs(optionsFromFile)
			if err != nil {
				return err
			}
		}
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_COMPILER_OPTIONS_FILE+"="+absOptionsFromFile)

		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		execCmdEnv = append(execCmdEnv, exec_tool.XGO_SRC_WD+"="+wd)
	}

	if logDebugFile != nil {
		// filter env
		var logEnv []string
		for _, env := range execCmdEnv {
			if !strings.HasPrefix(env, "GO") && !strings.HasPrefix(env, "XGO") {
				continue
			}
			logEnv = append(logEnv, env)
		}
		logDebug("command go env: %v...", logEnv)
	}

	printDir := projectDir
	if printDir == "" {
		printDir = "."
	}
	logDebug("command dir: %v", printDir)

	if projectDir != "" {
		execCmd.Dir = projectDir
	}
	execCmd.Env = execCmdEnv
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	logDebug("command executable path: %v", execCmd.Path)
	if logCmdExec != nil {
		logCmdExec()
	}
	err = execCmd.Run()
	if err != nil {
		return err
	}

	if debugMode {
		err := netutil.ServePort("localhost", 2345, true, 500*time.Millisecond, func(port int) {
			fmt.Fprintln(os.Stderr, debug_support.FormatDlvPrompt(port))
		}, func(port int) error {
			// dlv exec --api-version=2 --listen=localhost:2345 --accept-multiclient --headless ./debug.bin
			return cmd_support.Debug().Run("dlv", debug_support.FormatDlvArgs(finalBuildOutput, port, runFlagsAfterBuild)...)
		})
		if err != nil {
			return err
		}
	}

	// if dump IR is not nil, output to stdout
	if tmpIRFile != "" {
		// ast file not exists
		err := checkedCopyToStdout(tmpIRFile, "--dump-ir not effective, use -a to trigger recompile or check if you package path matches")
		if err != nil {
			return err
		}
	}
	if tmpASTFile != "" {
		// ast file not exists
		err := checkedCopyToStdout(tmpASTFile, "--dump-ast not effective, use -a to trigger recompile or check if you package path matches")
		if err != nil {
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

func checkedCopyToStdout(file string, msg string) error {
	err := copyToStdout(file)
	if err != nil {
		// ir file not exists
		if errors.Is(err, os.ErrNotExist) {
			return errors.New(msg)
		}
		return err
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
		if goVersion.Major != 1 || (minor < 17 || minor > 24) {
			return nil, fmt.Errorf("xgo only support go1.17 ~ go1.24, current: %s", goVersionStr)
		}
	}
	return goVersion, nil
}

// getOrMakeAbsXgoHome returns absolute path to xgo homoe
func getOrMakeAbsXgoHome(xgoHome string) (string, error) {
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

// getTmpDir for build caches
func getTmpDir() (string, error) {
	if runtime.GOOS != "windows" {
		// the publicly known /tmp dir
		const KNOWN_TMP = "/tmp"
		info, err := os.Stat(KNOWN_TMP)
		if err == nil && info.IsDir() {
			return KNOWN_TMP, nil
		}
	}
	tmpDir := os.TempDir()
	if tmpDir == "" {
		return "", fmt.Errorf("cannot get tmp dir")
	}
	return tmpDir, nil
}

func getNakedGo() string {
	if nakedGoReplacement != "" {
		return nakedGoReplacement
	}
	return "go"
}
func getGoEnvRoot(dir string) (string, error) {
	goroot, err := cmd.Dir(dir).Output(getNakedGo(), "env", "GOROOT")
	if err != nil {
		return "", err
	}
	// remove special characters from output
	goroot = strings.ReplaceAll(goroot, "\n", "")
	goroot = strings.ReplaceAll(goroot, "\r", "")
	return goroot, nil
}

var nakedGoReplacement = os.Getenv("XGO_REAL_GO_BINARY")

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

		var errSuffix string
		if envErr != nil {
			var errMsg string
			if e, ok := envErr.(*exec.ExitError); ok {
				errMsg = string(e.Stderr)
			} else {
				errMsg = envErr.Error()
			}
			errSuffix = fmt.Sprintf(": go env GOROOT: %v", errMsg)
		}
		return "", fmt.Errorf("requires GOROOT or --with-goroot%s", errSuffix)
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
	if V1_0_0 {
		// only needed for older xgo
		err = os.MkdirAll(packageDataDir, 0755)
		if err != nil {
			return fmt.Errorf("create %s: %w", filepath.Base(packageDataDir), err)
		}
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

func parseErr(err error) (code int, stderr []byte) {
	if err == nil {
		return 0, nil
	}
	if err, ok := err.(*exec.ExitError); ok {
		return err.ExitCode(), err.Stderr
	}
	return 1, []byte(err.Error())
}

func consumeErrAndExit(err error) {
	code, stderr := parseErr(err)
	if len(stderr) > 0 {
		os.Stderr.Write(stderr)
		os.Stderr.WriteString("\n")
	}
	os.Exit(code)
}
