package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/cmd/xgo/exec_tool"
)

// usage:
// go build -toolexec="go-tool-exec insert-trap some/pkg.Fn"
func main() {
	// os.Arg[0] = exec_tool
	// os.Arg[1] = compile or others...
	args := os.Args[1:]
	err := exec_tool.ToolExecMain(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
