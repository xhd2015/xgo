package main

import (
	"os"

	"github.com/xhd2015/xgo/cmd/xgo/coverage"
)

func main() {
	// os.Arg[0] = coverage
	// os.Arg[1] = args
	args := os.Args[1:]
	coverage.Main(args)
}
