package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Output(dir string, cmdName string, args ...string) (string, error) {
	out, err := OutputBytes(dir, cmdName, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

func OutputBytes(dir string, cmdName string, args ...string) ([]byte, error) {
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%s: %w\n%s", cmd, err, stderr.String())
	}
	return out, nil
}

func RunLogged(dir string, env []string, cmdName string, args ...string) error {
	Logf("+ %s", ShellQuoteArgs(cmdName, args))
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %w", cmd, err)
	}
	return nil
}

func ShellQuoteArgs(cmdName string, args []string) string {
	parts := append([]string{cmdName}, args...)
	for i, p := range parts {
		parts[i] = shellQuote(p)
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		if !(r == '/' || r == '.' || r == '-' || r == '_' || r == '=' || r == ':' ||
			('0' <= r && r <= '9') ||
			('A' <= r && r <= 'Z') ||
			('a' <= r && r <= 'z')) {
			return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		}
	}
	return s
}
