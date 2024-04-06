package main

import (
	"fmt"
	"os"
	"os/exec"
)

// usage:
//
//	go run ./script/setup-dev
//	go run ./script/setup-dev --with-goroot go1.20.14
func main() {
	args := os.Args[1:]
	execArgs := []string{
		"run",
		"-tags", "dev",
		"./cmd/xgo",
		"build",
		"--xgo-src",
		"./",
		"--setup-dev",
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
