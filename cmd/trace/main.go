package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/cmd/xgo/trace"
)

func main() {
	args := os.Args[1:]
	err := trace.Main(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
