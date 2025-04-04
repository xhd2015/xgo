//go:build go1.18
// +build go1.18

package inspect

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/internal/trap"
)

func GenericFunc[T any]() {
}

type GenericType[T any] struct {
}

func (g *GenericType[T]) GenericMethod() {
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
