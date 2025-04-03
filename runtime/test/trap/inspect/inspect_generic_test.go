//go:build go1.18
// +build go1.18

package inspect

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/trap"
)

func GenericFunc[T any]() {
}

type GenericType[T any] struct {
}

func (g *GenericType[T]) GenericMethod() {
}

func TestInspectGeneric222(t *testing.T) {
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
	if recv != nil {
		t.Fatalf("expect receive for GenericType[int].GenericMethod to be nil, actual: %v", recv)
	}
}
