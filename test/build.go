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

	noPipeStderr bool
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
	if opts == nil || !opts.noPipeStderr {
		cmd.Stderr = os.Stderr
	}
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

// return clean up func
type buildRuntimeOpts struct {
	xgoBuildEnv []string
	runEnv      []string
}

func buildWithRuntimeAndOutput(dir string, opts buildRuntimeOpts) (string, error) {
	tmpFile, err := getTempFile("test")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpFile)

	// func_list depends on xgo/runtime, but xgo/runtime is
	// a separate module, so we need to merge them
	// together first
	tmpDir, funcListDir, err := tmpMergeRuntimeAndTest(dir)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	_, err = xgoBuild([]string{
		"-o", tmpFile,
		"--project-dir", funcListDir,
		".",
	}, &options{
		env: opts.xgoBuildEnv,
	})
	if err != nil {
		return "", err
	}
	runCmd := exec.Command(tmpFile)
	runCmd.Env = os.Environ()
	runCmd.Env = append(runCmd.Env, opts.runEnv...)
	output, err := runCmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
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
