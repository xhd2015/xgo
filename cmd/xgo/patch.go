package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/filecopy"
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
type _FilePath []string

func (c _FilePath) Join(s ...string) string {
	return filepath.Join(filepath.Join(s...), filepath.Join(c...))
}

var affectedFiles []_FilePath

func init() {
	affectedFiles = append(affectedFiles, compilerFiles...)
	affectedFiles = append(affectedFiles, runtimeFiles...)
	affectedFiles = append(affectedFiles, reflectFiles...)
}

// assume go 1.20
// the patch should be idempotent
// the origGoroot is used to generate runtime defs, see https://github.com/xhd2015/xgo/issues/4#issuecomment-2017880791
func patchRuntimeAndCompiler(origGoroot string, goroot string, xgoSrc string, goVersion *goinfo.GoVersion, syncWithLink bool, revisionChanged bool) error {
	if goroot == "" {
		return fmt.Errorf("requires goroot")
	}
	if isDevelopment && xgoSrc == "" {
		return fmt.Errorf("requires xgoSrc")
	}
	if !isDevelopment && !revisionChanged {
		return nil
	}

	// runtime
	err := patchRuntimeAndTesting(goroot)
	if err != nil {
		return err
	}

	// compiler
	err = patchCompiler(origGoroot, goroot, goVersion, xgoSrc, revisionChanged, syncWithLink)
	if err != nil {
		return err
	}

	return nil
}

func replaceBuildIgnore(content []byte) ([]byte, error) {
	const buildIgnore = "//go:build ignore"

	// buggy: content = bytes.Replace(content, []byte("//go:build ignore\n"), nil, 1)
	return replaceMarkerNewline(content, []byte(buildIgnore))
}

// content = bytes.Replace(content, []byte("//go:build ignore\n"), nil, 1)
func replaceMarkerNewline(content []byte, marker []byte) ([]byte, error) {
	idx := bytes.Index(content, marker)
	if idx < 0 {
		return nil, fmt.Errorf("missing %s", string(marker))
	}
	idx += len(marker)
	if idx < len(content) && content[idx] == '\r' {
		idx++
	}
	if idx < len(content) && content[idx] == '\n' {
		idx++
	}
	return content[idx:], nil
}
func checkRevisionChanged(revisionFile string, currentRevision string) (bool, error) {
	savedRevision, err := readOrEmpty(revisionFile)
	if err != nil {
		return false, err
	}
	logDebug("current revision: %s, last revision: %s from file %s", currentRevision, savedRevision, revisionFile)
	if savedRevision == "" || savedRevision != currentRevision {
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

// NOTE: flagA never cause goroot to reset
func syncGoroot(goroot string, instrumentGoroot string, fullSyncRecordFile string) error {
	// check if src goroot has src/runtime
	srcRuntimeDir := filepath.Join(goroot, "src", "runtime")
	err := assertDir(srcRuntimeDir)
	if err != nil {
		return err
	}
	var goBinaryChanged bool = true
	srcGoBin := filepath.Join(goroot, "bin", "go")
	dstGoBin := filepath.Join(instrumentGoroot, "bin", "go")

	srcFile, err := os.Stat(srcGoBin)
	if err != nil {
		return nil
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

	if dstFile != nil && !dstFile.IsDir() && dstFile.Size() == srcFile.Size() {
		goBinaryChanged = false
	}

	var doPartialCopy bool
	if !goBinaryChanged && statNoErr(fullSyncRecordFile) {
		// full sync record does not yet exist
		doPartialCopy = true
	}
	if doPartialCopy {
		// do partial copy
		err := partialCopy(goroot, instrumentGoroot)
		if err != nil {
			return err
		}
	} else {
		rmErr := os.Remove(fullSyncRecordFile)
		if rmErr != nil {
			if !errors.Is(rmErr, os.ErrNotExist) {
				return rmErr
			}
		}

		// need copy, delete target dst dir first
		// TODO: use git worktree add if .git exists
		err = filecopy.NewOptions().
			Concurrent(2). // 10 is too much
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

func partialCopy(goroot string, instrumentGoroot string) error {
	err := os.RemoveAll(xgoRewriteInternal.Join(instrumentGoroot))
	if err != nil {
		return err
	}
	for _, affectedFile := range affectedFiles {
		srcFile := affectedFile.Join(goroot)
		dstFile := affectedFile.Join(instrumentGoroot)

		err := filecopy.CopyFileAll(srcFile, dstFile)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			// delete dstFile
			err := os.RemoveAll(dstFile)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func statNoErr(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}
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
				// build exec tool
				buildExecToolCmd := exec.Command("go", "build", "-o", execToolBin, "./exec_tool")
				buildExecToolCmd.Dir = filepath.Join(xgoSrc, "cmd")
				buildExecToolCmd.Stdout = os.Stdout
				buildExecToolCmd.Stderr = os.Stderr
				err = buildExecToolCmd.Run()
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
func buildCompiler(goroot string, output string) error {
	args := []string{"build"}
	if isDevelopment {
		args = append(args, "-gcflags=all=-N -l")
	}
	args = append(args, "-o", output, "./")
	cmd := exec.Command(filepath.Join(goroot, "bin", "go"), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env, err := patchEnvWithGoroot(os.Environ(), goroot)
	if err != nil {
		return err
	}
	cmd.Env = env
	cmd.Dir = filepath.Join(goroot, "src", "cmd", "compile")
	return cmd.Run()
}

func compareAndUpdateCompilerID(compilerFile string, compilerIDFile string) (changed bool, err error) {
	prevData, statErr := ioutil.ReadFile(compilerIDFile)
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
	data, err := exec.Command("go", "tool", "buildid", file).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}
