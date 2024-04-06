package patch

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

var a int = 123

func TestPatchVarTest(t *testing.T) {
	mock.Patch(&a, func() int {
		return 456
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched varaibel a to be %d, actual: %d", 456, b)
	}
}
