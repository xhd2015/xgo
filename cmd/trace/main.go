package main

import (
	"os"

	"github.com/xhd2015/xgo/cmd/xgo/trace"
)

func main() {
	args := os.Args[1:]
	trace.Main(args)
}
