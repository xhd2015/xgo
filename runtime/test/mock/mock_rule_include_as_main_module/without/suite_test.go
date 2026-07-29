package without

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/app"
	"github.com/xhd2015/xgo/runtime/test/mock/mock_rule_include_as_main_module/other"
)

// Baseline: main_module-only rules, no include-as.
//
// Important: third-party mock.Patch targets must go through an interface{}
// helper so static mock-ref analysis does NOT auto-instrument them. Direct
// mock.Patch(app.Hello, ...) would instrument app regardless of main_module
// rules and would not test include-as-main-module.

func TestLocalPatchOK(t *testing.T) {
	mock.Patch(LocalHello, func() string { return "mock-local" })
	if got := LocalHello(); got != "mock-local" {
		t.Fatalf("LocalHello: want mock-local, got %q", got)
	}
}

func TestAppPatchFails(t *testing.T) {
	expectPatchFail(t, app.Hello)
}

func TestOtherPatchFails(t *testing.T) {
	expectPatchFail(t, other.Hello)
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
