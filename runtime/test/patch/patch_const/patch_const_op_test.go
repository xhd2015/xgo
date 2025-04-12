package patch_const

import (
	"fmt"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/mock"
)

const A = 20 * time.Second

func TestPatchConstOp(t *testing.T) {
	// patch should error because registering were filtered
	var errMsg string
	func() {
		defer func() {
			if e := recover(); e != nil {
				errMsg = fmt.Sprint(e)
			}
		}()
		mock.PatchByName("github.com/xhd2015/xgo/runtime/test/patch_const", "A", func() time.Duration {
			return 10 * time.Second
		})
	}()

	expectErr := "failed to setup mock for: github.com/xhd2015/xgo/runtime/test/patch_const.A"
	if errMsg != expectErr {
		t.Fatalf("expect errMsg to be: %q, actual: %q", expectErr, errMsg)
	}

	a := A
	if a != 20*time.Second {
		t.Fatalf("expect patch A failed because current xgo does not resolve operation type, actual: a=%v, want: %v", a, 20*time.Second)
	}
}
