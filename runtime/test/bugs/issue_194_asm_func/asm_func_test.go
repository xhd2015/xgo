package asm_func

import (
	"testing"
)

// see bug https://github.com/xhd2015/xgo/issues/202
func TestAsmFunc(t *testing.T) {
	// no panic
	// don't really call it, go can omit
	// this function.
	// but prior to fixing this bug, xgo
	// registers it, so the linkers failed
	// AsmFunc()
}
