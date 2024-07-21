package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

const GO_MAJOR_1 = 1
const GO_VERSION_16 = 16 // unsupported
const GO_VERSION_17 = 17
const GO_VERSION_18 = 18
const GO_VERSION_19 = 19
const GO_VERSION_20 = 20
const GO_VERSION_21 = 21
const GO_VERSION_22 = 22
const GO_VERSION_23 = 23

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
	// when building the compiler, we want native
	// build, see https://github.com/xhd2015/xgo/issues/231
	cmd.Env = appendNativeBuildEnv(cmd.Env)
	cmd.Dir = filepath.Join(goroot, "src", "cmd", "compile")
	return cmd.Run()
}
func buildExecTool(dir string, execToolBin string) error {
	// build exec tool
	cmd := exec.Command(getNakedGo(), "build", "-o", execToolBin, "./exec_tool")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = appendNativeBuildEnv(os.Environ())
	return cmd.Run()
}

func appendNativeBuildEnv(env []string) []string {
	return append(env, "GOOS=", "GOARCH=")
}
