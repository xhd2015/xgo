package cmd

import (
	"os"
	"os/exec"
	"strings"
)

func Output(cmd string, args ...string) (string, error) {
	return cmdExec(cmd, args, false)
}
func Run(cmd string, args ...string) error {
	_, err := cmdExec(cmd, args, true)
	return err
}
func cmdExec(cmd string, args []string, pipeStdout bool) (string, error) {
	execCmd := exec.Command(cmd, args...)
	execCmd.Stderr = os.Stderr
	if pipeStdout {
		execCmd.Stdout = os.Stdout
		return "", execCmd.Run()
	}
	out, err := execCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}
