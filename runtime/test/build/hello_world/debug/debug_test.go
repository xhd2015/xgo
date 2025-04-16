package debug

import (
	"testing"
)

func exists(v interface{}) {}

var __xgo_debug = func(v interface{}) {}

var __xgo_register_00 = func(v interface{}) {}

func init() {
	__xgo_init_00()
}

//go:noinline
func __xgo_init_00() {
	__xgo_debug = exists
	__xgo_register_00 = __xgo_register_00
}

func TestHelloWorld(t *testing.T) {
	if false {
		__xgo_register_00(&t)
	}
	t.Log("hello world")
}
