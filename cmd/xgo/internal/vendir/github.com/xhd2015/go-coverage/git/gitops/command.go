package gitops

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/sh"
)

func RunCommand(dir string, withCommand func(commands []string) []string) (string, error) {
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
	res, _, err := sh.RunBashWithOpts(commands, sh.RunBashOptions{
		Verbose:    true,
		NeedStdOut: true,
	})
	if err != nil {
		return "", err
	}
	return res, nil
}

func splitLinesFilterEmpty(s string) []string {
	list := strings.Split(s, "\n")
	idx := 0
	for _, e := range list {
		if e != "" {
			list[idx] = e
			idx++
		}
	}
	return list[:idx]
}
