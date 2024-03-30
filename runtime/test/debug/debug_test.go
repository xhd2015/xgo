// debug test is a convienent package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/trap"
)

type struct_ struct {
}

func (c *struct_) F() {
	panic("should never call me")
}

func TestDebug(t *testing.T) {
	s := &struct_{}
	srecv, sf := trap.Inspect(s.F)
	if sf == nil {
		t.Fatalf("struct_.F not found")
	}
	if srecv == nil {
		t.Fatalf("expect receive for struct_.F not nil, actual: %v", srecv)
	}
}
