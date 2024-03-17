package test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type options struct {
	run    bool
	exec   bool
	noTrim bool
	env    []string

	noPipeStderr bool
}

func runXgo(args []string, opts *options) (string, error) {
	var xgoCmd string = "build"
	if opts != nil {
		if opts.run {
			xgoCmd = "run"
		} else if opts.exec {
			xgoCmd = "exec"
		}
	}
	xgoArgs := append([]string{
		"run", "../cmd/xgo",
		xgoCmd,
		"--xgo-src",
		"../",
		"--sync-with-link",
	}, args...)
	cmd := exec.Command("go", xgoArgs...)
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

func xgoExec(args ...string) (string, error) {
	return runXgo(args, &options{
		exec: true,
	})
}

// return clean up func
type buildRuntimeOpts struct {
	xgoBuildArgs []string
	xgoBuildEnv  []string
	runEnv       []string

	debug bool
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

	xgoBuildArgs := []string{
		"-o", tmpFile,
		// "-a",
		"--project-dir", funcListDir,
	}
	if opts.debug {
		xgoBuildArgs = append(xgoBuildArgs, "-gcflags=all=-N -l")
	}
	xgoBuildArgs = append(xgoBuildArgs, opts.xgoBuildArgs...)
	xgoBuildArgs = append(xgoBuildArgs, ".")
	_, err = runXgo(xgoBuildArgs, &options{
		env: opts.xgoBuildEnv,
	})
	if err != nil {
		return "", err
	}
	if opts.debug {
		fmt.Println(tmpFile)
		time.Sleep(10 * time.Minute)
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
		_, err = runXgo(buildArgs, nil)
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
