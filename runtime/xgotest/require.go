// Package xgotest provides helpers for tests that need xgo instrumentation.
package xgotest

import (
	"os"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

// TestingT is the subset of testing.TB used by Require.
type TestingT interface {
	Helper()
	Skip(args ...interface{})
	Fatal(args ...interface{})
}

// Require fails the test if the binary was not built by xgo. Set XGO_REQUIRE=skip
// to skip instead (for example, when running tests from an editor). The default
// and XGO_REQUIRE=fail fail the test. In an xgo-built binary Require does nothing.
func Require(t TestingT) {
	t.Helper()
	if core.Instrumented() {
		return
	}
	switch mode := os.Getenv("XGO_REQUIRE"); mode {
	case "skip":
		t.Skip(runtime.NotInstrumentedMessage)
	case "", "fail":
		t.Fatal(runtime.NotInstrumentedMessage)
	default:
		t.Fatal("xgo: invalid XGO_REQUIRE=" + mode + "; expected skip or fail; " + runtime.NotInstrumentedMessage)
	}
}
