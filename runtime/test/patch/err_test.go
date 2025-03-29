package patch

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestPatchShouldWorkWithErrReturn(t *testing.T) {
	mock.Patch(toErr, func(s string) error {
		return fmt.Errorf("mock: %v", s)
	})
	err := toErr("test")
	if err.Error() != "mock: test" {
		t.Fatalf("expect toErr() patched to be %q, actual: %q", "mock: test", err.Error())
	}
}

func TestPatchShouldWorkWithErrNilReturn(t *testing.T) {
	mock.Patch(toErr, func(s string) error {
		return nil
	})
	err := toErr("test")
	if err != nil {
		t.Fatalf("expect toErr() patched to be nil, actual: %v", err)
	}
}
