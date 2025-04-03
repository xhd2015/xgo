//go:build go1.20
// +build go1.20

package inspect

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/trap"
)

func TestInspectGenericSuccessWithGo120(t *testing.T) {
	recv, f, _, _ := trap.InspectPC(GenericFunc[int])
	if f == nil {
		t.Fatalf("GenericFunc[int] not found")
	}
	if recv != nil {
		t.Fatalf("expect receive for GenericFunc[int] to be nil, actual: %v", recv)
	}

	s := &GenericType[int]{}
	recv, f, _, _ = trap.InspectPC(s.GenericMethod)
	if f == nil {
		t.Fatalf("GenericType[int].GenericMethod not found")
	}
	if recv == nil {
		t.Fatalf("expect receive for GenericType[int].GenericMethod to not be nil")
	}
	recvPtr := reflect.ValueOf(recv).Elem().Pointer()
	wantPtr := uintptr(unsafe.Pointer(s))
	if recvPtr != wantPtr {
		t.Fatalf("expect receive for GenericType[int].GenericMethod to be %v, actual: %v", wantPtr, recvPtr)
	}
}
