//go:build go1.20
// +build go1.20

package type_alias

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

type ServiceGeneric[T any] struct {
	hello T
}

func (s *ServiceGeneric[T]) Greet(name string) string {
	return fmt.Sprintf("%v %s", s.hello, name)
}

type ServiceGenericString = ServiceGeneric[string]

var sgsPtr = &ServiceGenericString{hello: "hello"}

var sgs = ServiceGenericString{hello: "hello"}

func TestTypeAliasGenericPtr(t *testing.T) {
	res := sgsPtr.Greet("world")
	if res != "hello world" {
		t.Errorf("expect sgs.Greet() = %v, but got %v", "hello world", res)
	}
	mock.Patch(&sgsPtr, func() *ServiceGenericString {
		return &ServiceGenericString{hello: "mock"}
	})
	result := sgsPtr.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect sgs.Greet() = %v, but got %v", expected, result)
	}
}

// see https://github.com/xhd2015/xgo/issues/314
// sgs should be rewritten to pointer, not value
func TestTypeAliasGenericNonPtr(t *testing.T) {
	res := sgs.Greet("world")
	if res != "hello world" {
		t.Errorf("expect sgs.Greet() = %v, but got %v", "hello world", res)
	}
	mock.Patch(&sgs, func() *ServiceGenericString {
		return &ServiceGenericString{hello: "mock"}
	})
	result := sgs.Greet("world")
	expected := "mock world"
	if result != expected {
		t.Errorf("expect sgs.Greet() = %v, but got %v", expected, result)
	}
}
