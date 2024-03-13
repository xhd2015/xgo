package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
	run    bool
	noTrim bool
	env    []string
}

func xgoBuild(args []string, opts *options) (string, error) {
	var xgoCmd string = "build"
	if opts != nil {
		if opts.run {
			xgoCmd = "run"
		}
	}
	buildArgs := append([]string{
		"run", "../cmd/xgo",
		xgoCmd,
		"--xgo-src",
		"../",
		"--sync-with-link",
	}, args...)
	cmd := exec.Command("go", buildArgs...)
	cmd.Stderr = os.Stderr
	if opts != nil && len(opts.env) > 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, opts.env...)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	outStr := string(output)
	if opts == nil || !opts.noTrim {
		outStr = strings.TrimSuffix(outStr, "\n")
	}
	return outStr, nil
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

func fatalExecErr(t *testing.T, err error) {
	if err, ok := err.(*exec.ExitError); ok {
		t.Fatalf("%v", string(err.Stderr))
	}
	t.Fatalf("%v", err)
}

func buildAndRunOutput(program string) (output string, err error) {
	return buildAndRunOutputArgs([]string{program}, buildAndOutputOptions{})
}

type buildAndOutputOptions struct {
	build func(args []string) error
}

func buildAndRunOutputArgs(args []string, opts buildAndOutputOptions) (output string, err error) {
	testBin, err := getTempFile("test")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(testBin)
	buildArgs := []string{"-o", testBin}
	buildArgs = append(buildArgs, args...)
	if opts.build != nil {
		err = opts.build(buildArgs)
	} else {
		_, err = xgoBuild(buildArgs, nil)
	}
	if err != nil {
		return "", err
	}
	out, err := exec.Command(testBin).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
