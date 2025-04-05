package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/xgo/instrument/instrument_go"
	"github.com/xhd2015/xgo/instrument/instrument_runtime"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/osinfo"
)

// the _FilePath declared at toplevel
// serves as an item list where
// the runtime and compiler maybe
// affected.
// NOTE: do not remove files, always add files,
// these old files may exists in older version
// so can be cleared by newer xgo
type _FilePath = patch.FilePath

// assume go 1.20
// the patch should be idempotent
// the origGoroot is used to generate runtime defs, see https://github.com/xhd2015/xgo/issues/4#issuecomment-2017880791
func patchRuntime(origGoroot string, goroot string, xgoSrc string, goVersion *goinfo.GoVersion, syncWithLink bool, resetOrCoreRevisionChanged bool) error {
	if goroot == "" {
		return fmt.Errorf("requires goroot")
	}
	if isDevelopment && xgoSrc == "" {
		return fmt.Errorf("requires xgoSrc")
	}
	if !isDevelopment && !resetOrCoreRevisionChanged {
		return nil
	}

	// instrument go
	err := instrument_go.InstrumentGo(goroot, goVersion)
	if err != nil {
		return err
	}

	// instrument go tool cover
	err = instrument_go.InstrumentGoToolCover(goroot, goVersion)
	if err != nil {
		return err
	}

	// instrument runtime
	err = instrument_runtime.InstrumentRuntime(goroot, goVersion, instrument_runtime.InstrumentRuntimeOptions{
		Mode: instrument_runtime.InstrumentMode_ForceAndIgnoreMark,
	})
	if err != nil {
		return err
	}

	if V1_0_0 {
		// runtime
		err := patchRuntimeAndTesting(origGoroot, goroot, goVersion)
		if err != nil {
			return err
		}

		// patch compiler
		err = patchCompiler(origGoroot, goroot, goVersion, xgoSrc, resetOrCoreRevisionChanged, syncWithLink)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkRevisionChanged(coreRevisionFile string, currentCoreRevision string) (bool, error) {
	savedCoreRevision, err := readOrEmpty(coreRevisionFile)
	if err != nil {
		return false, err
	}
	logDebug("current core revision: %s, last core revision: %s from file %s", currentCoreRevision, savedCoreRevision, coreRevisionFile)
	if savedCoreRevision == "" || savedCoreRevision != currentCoreRevision {
		return true, nil
	}
	return false, nil
}

func readOrEmpty(file string) (string, error) {
	version, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	s := string(version)
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\r")
	return s, nil
}

// syncGoroot copies goroot to instrumentGoroot
// NOTE: flagA never cause goroot to reset
func syncGoroot(goroot string, instrumentGoroot string, fullSyncRecordFile string) error {
	// check if src goroot has src/runtime
	srcRuntimeDir := filepath.Join(goroot, "src", "runtime")
	err := assertDir(srcRuntimeDir)
	if err != nil {
		return err
	}
	srcGoBin := filepath.Join(goroot, "bin", "go"+osinfo.EXE_SUFFIX)
	dstGoBin := filepath.Join(instrumentGoroot, "bin", "go"+osinfo.EXE_SUFFIX)

	srcFile, err := os.Stat(srcGoBin)
	if err != nil {
		return err
	}
	if srcFile.IsDir() {
		return fmt.Errorf("bad goroot: %s", goroot)
	}

	dstFile, statErr := os.Stat(dstGoBin)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return statErr
		}
	}

	var dstGoCopied bool
	if dstFile != nil && !dstFile.IsDir() {
		dstGoCopied = true
	}

	var onlyCopySrc bool
	if dstGoCopied && statNoErr(fullSyncRecordFile) {
		// full sync record does not yet exist
		onlyCopySrc = true
	}
	if onlyCopySrc {
		// do partial copy
		logDebug("partial copy $GOROOT/src")
		err := copyGorootSrc(goroot, instrumentGoroot)
		if err != nil {
			return err
		}
	} else {
		logDebug("fully copy $GOROOT and write %s", fullSyncRecordFile)
		rmErr := os.Remove(fullSyncRecordFile)
		if rmErr != nil {
			if !errors.Is(rmErr, os.ErrNotExist) {
				return rmErr
			}
		}

		// need copy, delete target dst dir first
		// TODO: use git worktree add if .git exists
		err = filecopy.NewOptions().
			Concurrent(1). // 10 is too much
			CopyReplaceDir(goroot, instrumentGoroot)
		if err != nil {
			return err
		}

		// record this full sync
		copyTime := time.Now().Format("2006-01-02T15:04:05Z07:00")
		err = os.WriteFile(fullSyncRecordFile, []byte(copyTime), 0755)
		if err != nil {
			return err
		}
	}
	// change binary executable
	return nil
}

func copyGorootSrc(goroot string, instrumentGoroot string) error {
	return filecopy.CopyReplaceDir(filepath.Join(goroot, "src"), filepath.Join(instrumentGoroot, "src"), false)
}

func statNoErr(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}

// Deprecated:
func buildInstrumentTool(goroot string, xgoSrc string, compilerBin string, compilerBuildIDFile string, execToolBin string, xgoBin string, debugPkg string, logCompile bool, noSetup bool, debugWithDlv bool) (compilerChanged bool, toolExecFlag string, err error) {
	var execToolCmd []string
	if !noSetup {
		// build the instrumented compiler
		err = buildCompiler(goroot, compilerBin)
		if err != nil {
			return false, "", err
		}
		compilerChanged, err = compareAndUpdateCompilerID(compilerBin, compilerBuildIDFile)
		if err != nil {
			return false, "", err
		}

		if false {
			actualExecToolBin := execToolBin
			if isDevelopment {
				err := buildExecTool(filepath.Join(xgoSrc, "cmd"), execToolBin)
				if err != nil {
					return false, "", err
				}
			} else {
				actualExecToolBin, err = findBuiltExecTool()
				if err != nil {
					return false, "", err
				}
			}
			// unused any more
			execToolCmd = []string{actualExecToolBin, "--enable"}
			_ = execToolCmd
		}
	}
	execToolCmd = []string{xgoBin, "exec_tool", "--enable"}

	if logCompile {
		execToolCmd = append(execToolCmd, "--log-compile")
	}
	if debugPkg != "" {
		execToolCmd = append(execToolCmd, "--debug="+debugPkg)
	}
	if debugWithDlv {
		execToolCmd = append(execToolCmd, "--debug-with-dlv")
	}
	// always add trailing '--' to mark exec tool flags end
	execToolCmd = append(execToolCmd, "--")

	toolExecFlag = "-toolexec=" + strings.Join(execToolCmd, " ")
	return compilerChanged, toolExecFlag, nil
}

// find exec_tool, first try the same dir with xgo,
// but if that is not found, we can fallback to ~/.xgo/bin/exec_tool
// because exec_tool changes rarely, so it is safe to use
// an older version.
// we may add version to check if exec_tool is compatible
// Deprecated: xgo itself embed exec tool
func findBuiltExecTool() (string, error) {
	dirName := filepath.Dir(os.Args[0])
	absDirName, err := filepath.Abs(dirName)
	if err != nil {
		return "", err
	}
	exeSuffix := osinfo.EXE_SUFFIX
	execToolBin := filepath.Join(absDirName, "exec_tool"+exeSuffix)
	_, statErr := os.Stat(execToolBin)
	if statErr == nil {
		return execToolBin, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("exec_tool not found in %s", dirName)
	}
	execToolBin = filepath.Join(homeDir, ".xgo", "bin", "exec_tool"+exeSuffix)
	_, statErr = os.Stat(execToolBin)
	if statErr == nil {
		return execToolBin, nil
	}
	return "", fmt.Errorf("exec_tool not found in %s and ~/.xgo/bin", dirName)
}

func compareAndUpdateCompilerID(compilerFile string, compilerIDFile string) (changed bool, err error) {
	prevData, statErr := fileutil.ReadFile(compilerIDFile)
	if statErr != nil {
		if !errors.Is(statErr, os.ErrNotExist) {
			return false, statErr
		}
	}
	prevID := string(prevData)
	curID, err := getBuildID(compilerFile)
	if err != nil {
		return false, err
	}
	if prevID != "" && prevID == curID {
		return false, nil
	}
	err = ioutil.WriteFile(compilerIDFile, []byte(curID), 0755)
	if err != nil {
		return false, err
	}
	return true, nil
}

func getBuildID(file string) (string, error) {
	data, err := exec.Command(getNakedGo(), "tool", "buildid", file).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}
