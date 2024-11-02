package git

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-inspect/sh"
)

func RunCommand(dir string, withCommand func(commands []string) []string) (string, error) {
	return RunCommandWithOptions(dir, withCommand, nil)
}
func RunCommandWithOptions(dir string, withCommand func(commands []string) []string, withOptions func(opts *sh.RunBashOptions)) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("requires dir")
	}
	if withCommand == nil {
		return "", fmt.Errorf("requires commandFilter")
	}
	commands := []string{
		"set -e",
		fmt.Sprintf("cd %s", sh.Quote(dir)),
	}
	commands = withCommand(commands)
	opts := sh.RunBashOptions{}
	if withOptions != nil {
		withOptions(&opts)
	}
	opts.Verbose = true
	opts.NeedStdOut = true
	res, _, err := sh.RunBashWithOpts(commands, opts)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(res, "\n"), nil
}

func RunCommands(dir string, cmds ...string) (string, error) {
	return RunCommand(dir, func(commands []string) []string {
		return append(commands, cmds...)
	})
}

func RunCommandsWithOptions(dir string, argOpts sh.RunBashOptions, cmds ...string) (string, error) {
	return RunCommandWithOptions(dir, func(commands []string) []string {
		return append(commands, cmds...)
	}, func(opts *sh.RunBashOptions) {
		*opts = argOpts
	})
}

func splitLinesFilterEmpty(s string) []string {
	list := strings.Split(s, "\n")
	idx := 0
	for _, e := range list {
		e = strings.TrimSpace(e)
		if e != "" {
			list[idx] = e
			idx++
		}
	}
	return list[:idx]
}
