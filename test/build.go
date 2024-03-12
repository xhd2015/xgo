package test

import (
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"github.com/xhd2015/xgo/support/filecopy"
)

func getTempFile(pattern string) (string, error) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return "", err
	}

	return filepath.Join(tmpDir, pattern), nil
}

type options struct {
}

func xgoBuild(args []string, opts *options) (string, error) {
	buildArgs := append([]string{
		"run", "../cmd/xgo",
		"build",
		"--xgo-src",
		"../",
		"--sync-with-link",
	}, args...)
	cmd := exec.Command("go", buildArgs...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(output), "\n"), nil
}

func tmpMergeRuntimeAndTest(testDir string) (rootDir string, subDir string, err error) {
	return linkRuntimeAndTest(testDir, false)
}

func tmpRuntimeModeAndTest(testDir string) (rootDir string, subDir string, err error) {
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
