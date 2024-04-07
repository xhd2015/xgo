// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

const N = 50
const M = 20

func TestPatchConstOperationShouldCompileAndSkipMock(t *testing.T) {
	// should have no effect
	mock.PatchByName("github.com/xhd2015/xgo/runtime/test/debug", "N", func() int {
		return 10
	})
	// because N is used inside an operation
	// it's type is not yet determined, so
	// should not rewrite it
	var size int64 = M + N
	t.Logf("size=%d", size)
	// size := (N + 1) * unsafe.Sizeof(int(0))
	if size != 11 {
		t.Fatalf("expect N not patched and size to be %d, actual: %d\n", 11, size)
	}
}
