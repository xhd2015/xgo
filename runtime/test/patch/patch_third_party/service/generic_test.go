//go:build go1.20
// +build go1.20

package service

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party/third/generic"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party/third/generic_fn"
)

func TestPatchGenericFunc(t *testing.T) {
	mock.Patch(generic_fn.Greet[string], func(name string) string {
		return "mock " + name
	})
	res := generic_fn.Greet("world")
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}

func TestPatchGenericServiceInstance(t *testing.T) {
	genericSvc := generic.GreetService[string]{}
	mock.Patch(genericSvc.Greet, func(name string) string {
		return "mock " + name
	})
	res := genericSvc.Greet("world")
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}
