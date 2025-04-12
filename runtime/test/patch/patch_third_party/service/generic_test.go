//go:build go1.20
// +build go1.20

package service

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party/third/generic"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_third_party/third/generic_fn"
)

func TestPatchGenericFuncSingle(t *testing.T) {
	mock.Patch(generic_fn.Greet[string], func(name string) string {
		return "mock " + name
	})
	res := generic_fn.Greet("world")
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}

func TestPatchGenericFuncMulti(t *testing.T) {
	mock.Patch(generic_fn.GreetMulti[string, string], func(h string, w string) string {
		return "mock " + h + " " + w
	})
	res := generic_fn.GreetMulti("hello", "world")
	if res != "mock hello world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock hello world", res)
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

func TestPatchGenericServiceInstanceMulti(t *testing.T) {
	genericSvc := generic.GreetMultiService[string, string]{}
	mock.Patch(genericSvc.GreetMulti, func(h string, w string) string {
		return "mock " + h + " " + w
	})
	res := genericSvc.GreetMulti("hello", "world")
	if res != "mock hello world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock hello world", res)
	}
}
