package flag_overrides_env

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/app"
	"github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/lib"
)

// Flag list is app; env list is lib. Flag must win the whole list (not merge):
// app mockable via main_module, lib not.

func TestLocalPatchOK(t *testing.T) {
	mock.Patch(LocalHello, func() string { return "mock-local" })
	if got := LocalHello(); got != "mock-local" {
		t.Fatalf("LocalHello: want mock-local, got %q", got)
	}
}

func TestAppPatchOK(t *testing.T) {
	expectPatchOK(t, app.Hello, app.Hello, "mock-app")
}

func TestLibPatchFails(t *testing.T) {
	expectPatchFail(t, lib.Hello)
}

func expectPatchOK(t *testing.T, fn interface{}, call func() string, want string) {
	t.Helper()
	var panicErr interface{}
	func() {
		defer func() { panicErr = recover() }()
		mock.Patch(fn, func() string { return want })
	}()
	if panicErr != nil {
		t.Fatalf("expect mock.Patch to succeed (include-as-main-module should instrument target), panic: %v", panicErr)
	}
	if got := call(); got != want {
		t.Fatalf("patched call: want %q, got %q", want, got)
	}
}

func expectPatchFail(t *testing.T, fn interface{}) {
	t.Helper()
	var panicErr interface{}
	func() {
		defer func() { panicErr = recover() }()
		mock.Patch(fn, func() string { return "should-not-apply" })
	}()
	if panicErr == nil {
		t.Fatalf("expect mock.Patch to panic for non-instrumented target, actual nil")
	}
	t.Logf("got expected panic: %s", fmt.Sprint(panicErr))
}
