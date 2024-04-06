package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	args := os.Args[1:]
	execArgs := []string{
		"run",
		"-tags", "dev",
		"./cmd/xgo",
		"build",
		"--xgo-src",
		"./",
		"--build-compiler",
	}
	execArgs = append(execArgs, args...)

	cmd := exec.Command("go", execArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
