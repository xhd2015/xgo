package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

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
	tmpDir, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		return "", "", err
	}

	// copy runtime to a tmp directory, and
	// test under there
	if goModOnly {
		err = filecopy.LinkFile("../runtime/go.mod", filepath.Join(tmpDir, "go.mod"))
	} else {
		err = filecopy.LinkFiles("../runtime", tmpDir)
	}
	if err != nil {
		return "", "", err
	}

	subDir, err = os.MkdirTemp(tmpDir, filepath.Base(testDir))
	if err != nil {
		return "", "", err
	}

	err = filecopy.LinkFiles(testDir, subDir)
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
var xgoInitOnce sync.Once

func ensureXgoInit() error {
	xgoInitOnce.Do(func() {
		_, xgoInitErr = runXgo([]string{"--build-compiler"}, &options{
			init: true,
		})
	})
	return xgoInitErr
}
