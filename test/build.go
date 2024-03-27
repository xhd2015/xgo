package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/xhd2015/xgo/support/osinfo"

	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/goinfo"
)

func getTempFile(pattern string) (string, error) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return "", err
	}

	return filepath.Join(tmpDir, pattern), nil
}

func tmpMergeRuntimeAndTest(testDir string) (rootDir string, subDir string, err error) {
	return linkRuntimeAndTest(testDir, false)
}

func tmpWithRuntimeGoModeAndTest(testDir string) (rootDir string, subDir string, err error) {
	return linkRuntimeAndTest(testDir, true)
}

func linkRuntimeAndTest(testDir string, goModOnly bool) (rootDir string, subDir string, err error) {
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		return "", "", err
	}

	// windows no link
	if runtime.GOOS != "windows" {
		// copy runtime to a tmp directory, and
		// test under there
		if goModOnly {
			err = filecopy.LinkFile(filepath.Join("..", "runtime", "go.mod"), filepath.Join(tmpDir, "go.mod"))
		} else {
			err = filecopy.LinkFiles(filepath.Join("..", "runtime"), tmpDir)
		}
	} else {
		if goModOnly {
			err = filecopy.CopyFile(filepath.Join("..", "runtime", "go.mod"), filepath.Join(tmpDir, "go.mod"))
		} else {
			err = filecopy.CopyReplaceDir(filepath.Join("..", "runtime"), tmpDir, false)
		}
	}

	if err != nil {
		return "", "", err
	}

	subDir, err = os.MkdirTemp(tmpDir, filepath.Base(testDir))
	if err != nil {
		return "", "", err
	}

	if runtime.GOOS != "windows" {
		err = filecopy.LinkFiles(testDir, subDir)
	} else {
		err = filecopy.CopyReplaceDir(testDir, subDir, false)
	}
	if err != nil {
		return "", "", err
	}
	return tmpDir, subDir, nil
}

func fatalExecErr(t *testing.T, err error) {
	if err, ok := err.(*exec.ExitError); ok {
		t.Fatalf("%v", string(err.Stderr))
	}
	t.Fatalf("%v", err)
}

func getErrMsg(err error) string {
	if err, ok := err.(*exec.ExitError); ok {
		return string(err.Stderr)
	}
	return err.Error()
}

var goVersionCached *goinfo.GoVersion
var goVersionErr error
var goVersionOnce sync.Once

func getGoVersion() (*goinfo.GoVersion, error) {
	goVersionOnce.Do(func() {
		goVersionStr, err := goinfo.GetGoVersionOutput("go")
		if err != nil {
			goVersionErr = err
			return
		}
		goVersionCached, goVersionErr = goinfo.ParseGoVersion(goVersionStr)
	})
	return goVersionCached, goVersionErr
}

var xgoInitErr error

var xgoBinary string

func init() {
	curDir, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	exeSuffix := osinfo.EXE_SUFFIX
	xgoBinary = filepath.Join(curDir, "xgo"+exeSuffix)
}

var xgoInitOnce sync.Once

const logs = false

// const logs = true

func ensureXgoInit() error {
	xgoInitOnce.Do(func() {
		if logs {
			fmt.Printf("init xgo\n")
		}
		_, xgoInitErr = runXgo([]string{"--build-compiler"}, &options{
			init: true,
		})
	})
	return xgoInitErr
}
