package cmd

import (
	"os"
	"os/exec"
	"strings"
)

func Output(cmd string, args ...string) (string, error) {
	return cmdExec(cmd, args, "", false)
}
func Run(cmd string, args ...string) error {
	_, err := cmdExec(cmd, args, "", true)
	return err
}

type CmdBuilder struct {
	env []string
	dir string
}

func Env(env []string) *CmdBuilder {
	return &CmdBuilder{
		env: env,
	}
}
func Dir(dir string) *CmdBuilder {
	return &CmdBuilder{
		dir: dir,
	}
}

func (c *CmdBuilder) Output(cmd string, args ...string) (string, error) {
	return cmdExecEnv(cmd, args, c.env, c.dir, false)
}
func (c *CmdBuilder) Run(cmd string, args ...string) error {
	_, err := cmdExecEnv(cmd, args, c.env, c.dir, false)
	return err
}

func cmdExec(cmd string, args []string, dir string, pipeStdout bool) (string, error) {
	return cmdExecEnv(cmd, args, nil, dir, pipeStdout)
}
func cmdExecEnv(cmd string, args []string, env []string, dir string, pipeStdout bool) (string, error) {
	execCmd := exec.Command(cmd, args...)
	execCmd.Stderr = os.Stderr
	if pipeStdout {
		execCmd.Stdout = os.Stdout
		return "", execCmd.Run()
	}
	if len(env) > 0 {
		execCmd.Env = os.Environ()
		execCmd.Env = append(execCmd.Env, env...)
	}
	execCmd.Dir = dir
	out, err := execCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}
