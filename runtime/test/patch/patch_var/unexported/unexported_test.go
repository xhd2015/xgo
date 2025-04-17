package unexported

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/test/patch/patch_var/unexported/pkg"
)

var A = pkg.NewUnexported(1)

func TestUnexported(t *testing.T) {
	a := A.Get()
	if a != 1 {
		t.Fatalf("a is not 1")
	}
	// cannot patch A
}
