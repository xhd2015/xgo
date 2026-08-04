package external_overlay_instrument

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// Caller -overlay replaces subject.go; xgo must instrument that replacement so
// mock.Patch works (proves A→B→C composition, not a bare A→B file redirect).
func TestExternalOverlayInstrumentComposesAndMocks(t *testing.T) {
	if got, want := Value(), "caller replacement"; got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}

	mock.Patch(Value, func() string { return "mocked" })
	if got, want := Value(), "mocked"; got != want {
		t.Fatalf("after mock Value() = %q, want %q (instrumentation missing on overlaid file?)", got, want)
	}
}
