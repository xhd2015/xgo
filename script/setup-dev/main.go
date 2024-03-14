package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run",
		"./cmd/xgo",
		"build",
		"--xgo-src",
		"./",
		"--setup-dev",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
