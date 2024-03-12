package main

import (
	"github.com/xhd2015/xgo/runtime/core/trace"
	"github.com/xhd2015/xgo/runtime/core/trap"
)

func init() {
	trap.Use()
	trace.Use()
}

func main() {
	ReadAtLeast("hello ")
	ReadAtLeast("trap")
}

func ReadAtLeast(a string) {
	print(a)
}
