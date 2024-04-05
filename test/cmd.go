package test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/osinfo"

	"github.com/xhd2015/xgo/support/cmd"
)

type xgoCmd string

const (
	xgoCmd_build     xgoCmd = "build"
	xgoCmd_run       xgoCmd = "run"
	xgoCmd_test      xgoCmd = "test"
	xgoCmd_testBuild xgoCmd = "test_build"
)

type options struct {
	xgoCmd xgoCmd
	exec   bool
	noTrim bool
	env    []string

	noPipeStderr bool

	stderr io.Writer

	init bool

	buildArgs  []string
	projectDir string
}

func runXgo(args []string, opts *options) (string, error) {
	if logs {
		fmt.Printf("run xgo: args=%v, opts=%+v\n", args, opts)
	}
	if opts == nil || !opts.init {
		err := ensureXgoInit()
		if err != nil {
			return "", err
		}
	} else {
		if logs {
			fmt.Printf("build xgo binary: output=%v\n", xgoBinary)
		}
		// build the xgo binary
		err := cmd.Run("go", "build", "-tags", "dev", "-o", xgoBinary, "../cmd/xgo")
		if err != nil {
			return "", err
		}
	}
	var xgoCmd string = "build"
	var extraArgs []string
	if opts != nil {
		if opts.xgoCmd == xgoCmd_build {
			xgoCmd = "build"
		} else if opts.xgoCmd == xgoCmd_run {
			xgoCmd = "run"
		} else if opts.xgoCmd == xgoCmd_test {
			xgoCmd = "test"
		} else if opts.xgoCmd == xgoCmd_testBuild {
			xgoCmd = "test"
			extraArgs = append(extraArgs, "-c")
		} else if opts.exec {
			xgoCmd = "exec"
		}
		extraArgs = append(extraArgs, opts.buildArgs...)
	}
	xgoArgs := []string{
		xgoCmd,
		"--xgo-src",
		"../",
		"--sync-with-link",
	}
	xgoArgs = append(xgoArgs, extraArgs...)
	// accerlate
	if opts == nil || !opts.init {
		xgoArgs = append(xgoArgs, "--no-setup")
	}
	if opts != nil && opts.projectDir != "" {
		xgoArgs = append(xgoArgs, "--project-dir", opts.projectDir)
	}
	xgoArgs = append(xgoArgs, args...)

	cmd := exec.Command(xgoBinary, xgoArgs...)
	if opts == nil {
		cmd.Stderr = os.Stderr
	} else if opts.stderr != nil {
		cmd.Stderr = opts.stderr
	} else if !opts.noPipeStderr {
		cmd.Stderr = os.Stderr
	}
	if opts != nil && len(opts.env) > 0 {
		cmd.Env = os.Environ()
		if logs {
			fmt.Printf("adding xgo env: %v\n", opts.env)
		}
		cmd.Env = append(cmd.Env, opts.env...)
	}
	if logs {
		fmt.Printf("executing xgo: %v\n", xgoArgs)
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
	exeSuffix := osinfo.EXE_SUFFIX
	tmpFile += exeSuffix
	defer os.RemoveAll(tmpFile)

	tmpExe := tmpFile
	// func_list depends on xgo/runtime, but xgo/runtime is
	// a separate module, so we need to merge them
	// together first
	tmpDir, funcListDir, err := tmpMergeRuntimeAndTest(dir)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	xgoBuildArgs := []string{
		"-o", tmpExe,
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
		fmt.Println(tmpExe)
		time.Sleep(10 * time.Minute)
	}
	runCmd := exec.Command(tmpExe)
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

	buildTest  bool
	buildArgs  []string
	projectDir string
}

func buildAndRunOutputArgs(args []string, opts buildAndOutputOptions) (output string, err error) {
	testBin, err := getTempFile("test")
	if err != nil {
		return "", err
	}
	exeSuffix := osinfo.EXE_SUFFIX
	testBin += exeSuffix
	defer os.RemoveAll(testBin)

	buildArgs := []string{"-o", testBin}
	buildArgs = append(buildArgs, args...)
	if opts.build != nil {
		err = opts.build(buildArgs)
	} else {
		cmd := xgoCmd_build
		if opts.buildTest {
			cmd = xgoCmd_testBuild
		}
		_, err = runXgo(buildArgs, &options{
			xgoCmd: cmd,

			buildArgs:  opts.buildArgs,
			projectDir: opts.projectDir,
		})
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
