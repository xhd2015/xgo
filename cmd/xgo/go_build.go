package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Deprecated: since xgo v1.1.0, xgo does not instrument the compiler anymore
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
