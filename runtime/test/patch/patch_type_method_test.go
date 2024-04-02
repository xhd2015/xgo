package patch

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestPatchTypeMethod(t *testing.T) {
	ins := &struct_{
		s: "world",
	}
	mock.Patch((*struct_).greet, func(ins *struct_) string {
		return "mock " + ins.s
	})

	res := ins.greet()
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}
