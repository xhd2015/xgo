package external_overlay_composition

import "testing"

// The caller overlay replaces subject.go before xgo instruments it. The xgo
// output overlay must therefore compose original -> caller replacement ->
// instrumented replacement, preserving the caller-visible behavior.
func TestExternalOverlayIsComposedAfterInstrumentation(t *testing.T) {
	if got, want := Value(), "caller replacement"; got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}
