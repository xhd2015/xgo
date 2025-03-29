package cmd

import (
	"fmt"
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
	env         []string
	dir         string
	debug       bool
	ignoreError bool
	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer
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
func Debug() *CmdBuilder {
	return &CmdBuilder{
		debug: true,
	}
}
func IgnoreError(b ...bool) *CmdBuilder {
	c := &CmdBuilder{}
	return c.IgnoreError(b...)
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
func (c *CmdBuilder) Stdin(stdin io.Reader) *CmdBuilder {
	c.stdin = stdin
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

func (c *CmdBuilder) Debug() *CmdBuilder {
	c.debug = true
	return c
}

func (c *CmdBuilder) IgnoreError(b ...bool) *CmdBuilder {
	ignore := true
	if len(b) > 0 {
		ignore = b[0]
	}
	c.ignoreError = ignore
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
	var stderr io.Writer
	if c != nil && c.stderr != nil {
		stderr = c.stderr
	} else {
		stderr = os.Stderr
	}
	if c != nil && c.debug {
		var lines []string
		if len(env) > 0 {
			for _, e := range env {
				lines = append(lines, "# "+e)
			}
		}
		if dir != "" {
			lines = append(lines, "# cd "+Quote(dir))
		}
		cmdQuotes := make([]string, 0, 1+len(args))
		cmdQuotes = append(cmdQuotes, cmd)
		for _, arg := range args {
			cmdQuotes = append(cmdQuotes, Quote(arg))
		}
		cmdStr := strings.Join(cmdQuotes, " ")
		lines = append(lines, cmdStr)
		for _, line := range lines {
			fmt.Fprintln(stderr, line)
		}
	}

	execCmd := exec.Command(cmd, args...)
	execCmd.Stderr = stderr
	if c != nil {
		execCmd.Stdin = c.stdin
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
	outStr := string(out)
	outStr = strings.TrimSuffix(outStr, "\r")
	outStr = strings.TrimSuffix(outStr, "\n")
	if err != nil {
		if c != nil && !c.ignoreError {
			return outStr, err
		}
		err = nil
	}
	return outStr, nil
}
