package patch_const

import (
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/mock"
)

const A = 20 * time.Second

func TestPatchConstOp(t *testing.T) {
	mock.PatchByName("github.com/xhd2015/xgo/runtime/test/patch_const", "A", func() time.Duration {
		return 10 * time.Second
	})
	a := A
	if a != 20*time.Second {
		t.Fatalf("expect patch A failed because current xgo does not resolve operation type, actual: a=%v, want: %v", a, 20*time.Second)
	}
}
