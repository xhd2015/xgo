//go:build ignore
// +build ignore

package check_trace_flag

import (
	"testing"

	_ "github.com/xhd2015/xgo/runtime/trace"
)

// check if 'xgo test --strace' would actually work
// it should generate a trace next the test file
// without explicit `defer trace.Begin()()`,
// even if panic
func TestCheckTraceFlagNormal(t *testing.T) {
	A()
}

func TestCheckTraceFlagPanic(t *testing.T) {
	A()
	panic("test")
}

func A() { B() }
func B() {}
