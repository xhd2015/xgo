package main

import (
	"fmt"
	"os"

	test_explorer "github.com/xhd2015/xgo/cmd/xgo/test-explorer"
)

func main() {
	err := test_explorer.Main(os.Args[1:], nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
