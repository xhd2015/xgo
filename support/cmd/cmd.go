package cmd

import (
	"io"
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
	env    []string
	dir    string
	stdout io.Writer
	stderr io.Writer
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

func New() *CmdBuilder {
	return &CmdBuilder{}
}

func (c *CmdBuilder) Env(env []string) *CmdBuilder {
	c.env = env
	return c
}
func (c *CmdBuilder) Dir(dir string) *CmdBuilder {
	c.dir = dir
	return c
}

func (c *CmdBuilder) Stdout(stdout io.Writer) *CmdBuilder {
	c.stdout = stdout
	return c
}

func (c *CmdBuilder) Stderr(stderr io.Writer) *CmdBuilder {
	c.stderr = stderr
	return c
}

func (c *CmdBuilder) Output(cmd string, args ...string) (string, error) {
	return cmdExecEnv(cmd, args, c.env, c.dir, false, c)
}

func (c *CmdBuilder) Run(cmd string, args ...string) error {
	_, err := cmdExecEnv(cmd, args, c.env, c.dir, true, c)
	return err
}

func cmdExec(cmd string, args []string, dir string, pipeStdout bool) (string, error) {
	return cmdExecEnv(cmd, args, nil, dir, pipeStdout, nil)
}
func cmdExecEnv(cmd string, args []string, env []string, dir string, useStdout bool, c *CmdBuilder) (string, error) {
	execCmd := exec.Command(cmd, args...)
	if c != nil && c.stderr != nil {
		execCmd.Stderr = c.stderr
	} else {
		execCmd.Stderr = os.Stderr
	}
	if len(env) > 0 {
		execCmd.Env = os.Environ()
		execCmd.Env = append(execCmd.Env, env...)
	}
	execCmd.Dir = dir
	if c != nil && c.stdout != nil {
		execCmd.Stdout = c.stdout
		return "", execCmd.Run()
	} else if useStdout {
		execCmd.Stdout = os.Stdout
		return "", execCmd.Run()
	}
	out, err := execCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}
