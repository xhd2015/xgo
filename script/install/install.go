package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/xhd2015/xgo/script/install/upgrade"
)

func main() {
	err := upgrade.Upgrade("")
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "%s\n", string(e.Stderr))
			os.Exit(e.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
