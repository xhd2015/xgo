//go:build ignore

package debug

import (
	"testing"
)

var __xgo_trap_0 = func() {}

func init() {
	__xgo_init_0()
}

//go:noinline
func __xgo_init_0() {
	__xgo_trap_0 = __xgo_trap_0
}

func TestHelloWorld(t *testing.T) {
	__xgo_trap_0()
	t.Log("hello world")
}
