//go:build go1.20
// +build go1.20

package inspect

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/internal/trap"
)

func TestInspectGenericSuccessWithGo120(t *testing.T) {
	recv, f, _, _ := trap.Inspect(GenericFunc[int])
	if f == nil {
		t.Fatalf("GenericFunc[int] not found")
	}
	if recv != nil {
		t.Fatalf("expect receive for GenericFunc[int] to be nil, actual: %v", recv)
	}

	s := &GenericType[int]{}
	recv, f, _, _ = trap.Inspect(s.GenericMethod)
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

func TestInspectPCGenericGo120ShouldSuccess(t *testing.T) {
	_, fnInfo, funcPC, trappingPC := trap.Inspect(ToString[int])
	_, fnInfoStr, funcPCStr, trappingPCStr := trap.Inspect(ToString[string])

	if !fnInfo.Generic {
		t.Fatalf("expect ToString[int] to be generic")
	}

	if false {
		// debug
		t.Logf("funcPC: 0x%x", funcPC)
		t.Logf("trappingPC: 0x%x", trappingPC)
		t.Logf("funcPCStr: 0x%x", funcPCStr)
		t.Logf("trappingPCStr: 0x%x", trappingPCStr)
	}

	if fnInfo != fnInfoStr {
		t.Fatalf("expect ToString[int] and ToString[string] returns the same func info")
	}

	if funcPC == trappingPC {
		t.Fatalf("expect ToString[int] pc to be different, actually the same: funcPC=0x%x, trappingPC=0x%x", funcPC, trappingPC)
	}
	if funcPCStr == trappingPCStr {
		t.Fatalf("expect ToString[string] pc to be different, actually the same: funcPC=0x%x, trappingPC=0x%x", funcPCStr, trappingPCStr)
	}
	if funcPC == funcPCStr {
		t.Fatalf("expect ToString[int] and ToString[string] to have different pc, actually the same: funcPC=0x%x, funcPCStr=0x%x", funcPC, funcPCStr)
	}
	if trappingPC == trappingPCStr {
		t.Fatalf("expect ToString[int] and ToString[string] to have different trapping PC, actually the same: trappingPC=0x%x, trappingPCStr=0x%x", trappingPC, trappingPCStr)
	}
}

func TestInspectGenericFunc(t *testing.T) {
	recvInt, fInt, fnPCInt, trappingPCInt := trap.Inspect(GenericFunc[int])
	if fInt == nil {
		t.Fatalf("GenericFunc[int] not found")
	}
	if recvInt != nil {
		t.Fatalf("expect receive for GenericFunc[int] to be nil, actual: %v", recvInt)
	}

	recvStr, fStr, fnPCStr, trappingPCStr := trap.Inspect(GenericFunc[string])
	if fStr == nil {
		t.Fatalf("GenericFunc[string] not found")
	}
	if recvStr != nil {
		t.Fatalf("expect receive for GenericFunc[string] to be nil, actual: %v", recvInt)
	}

	if fStr != fInt {
		t.Errorf("expect fInt==fStr, actual not same")
	}
	if fnPCInt == fnPCStr {
		t.Errorf("expect gfpcInt!=gfpcStr, actual same")
	}
	if trappingPCInt == trappingPCStr {
		t.Errorf("expect gtpcInt!=gtpcStr, actual same")
	}
}

func TestInspectGenericType(t *testing.T) {
	s := &GenericType[int]{}
	recvMethod, fMethod, _, _ := trap.Inspect(s.GenericMethod)
	if fMethod == nil {
		t.Fatalf("GenericType[int].GenericMethod not found")
	}
	if recvMethod == nil {
		t.Fatalf("expect receive for GenericType[int].GenericMethod to be not nil, actual: %v", recvMethod)
	}
}
