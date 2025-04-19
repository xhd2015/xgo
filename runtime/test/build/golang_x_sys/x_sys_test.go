//go:build !windows
// +build !windows

package golang_x_sys

import (
	"reflect"
	"testing"

	"golang.org/x/sys/unix"
)

func TestXSys(t *testing.T) {
	v := unix.Syscall
	p := reflect.ValueOf(v).Pointer()

	if p == 0 {
		t.Errorf("expect p to be non-zero, actual: %d", p)
	}
}
