//go:build go1.20
// +build go1.20

package trap_guard

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// mocking generic functions needs go1.20
func TestPatchGenericAndVariadic(t *testing.T) {
	mock.Patch(First[string], func(items ...string) (string, bool) {
		return "generic", true
	})
	mock.Patch(Join, func(sep string, parts ...string) string {
		return "joined " + sep
	})
	if v, ok := First("a"); !ok || v != "generic" {
		t.Errorf("First: %q %v", v, ok)
	}
	if v, ok := First(1); !ok || v != 1 {
		t.Errorf("First[int] must not be mocked: %v %v", v, ok)
	}
	if got := Join("-", "a"); got != "joined -" {
		t.Errorf("Join: %q", got)
	}
}
