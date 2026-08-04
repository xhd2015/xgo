package external_overlay_composition

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// The caller overlay replaces subject.go before xgo instruments it. The xgo
// output overlay must therefore compose original -> caller replacement ->
// instrumented replacement, preserving the caller-visible behavior and
// allowing traps/mocks on the overlaid function.
func TestExternalOverlayIsComposedAfterInstrumentation(t *testing.T) {
	if got, want := Value(), "caller replacement"; got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}

	// Instrumentation must apply to the overlaid replacement (B→C), not only
	// leave a plain file redirect A→B. Without traps, mock.Patch is a no-op.
	mock.Patch(Value, func() string { return "mocked" })
	if got, want := Value(), "mocked"; got != want {
		t.Fatalf("after mock Value() = %q, want %q (overlay composition dropped instrumentation?)", got, want)
	}
}
