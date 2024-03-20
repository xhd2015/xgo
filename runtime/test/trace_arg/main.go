package main

import (
	"github.com/xhd2015/xgo/runtime/trace"
)

func init() {
	trace.Enable()
}

func main() {
	ReadAtLeast("hello ")
	ReadAtLeast("trap")
}

func ReadAtLeast(a string) {
	print(a)
}

func ReadAtLeast_trap(a string) {
	after, stop := __xgo_trap(nil, []interface{}{&a}, []interface{}{})
	if stop {
	} else {
		if after != nil {
			defer after()
		}
		print(a)
	}
}

//go:noinline
func __xgo_trap(recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	return nil, false
}
