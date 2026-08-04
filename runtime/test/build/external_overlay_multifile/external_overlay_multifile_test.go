package external_overlay_multifile

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestMultiFileOverlayComposesAndMocks(t *testing.T) {
	if got, want := ValueA(), "replacement-a"; got != want {
		t.Fatalf("ValueA() = %q, want %q", got, want)
	}
	if got, want := ValueB(), "replacement-b"; got != want {
		t.Fatalf("ValueB() = %q, want %q", got, want)
	}

	mock.Patch(ValueA, func() string { return "mock-a" })
	mock.Patch(ValueB, func() string { return "mock-b" })
	if got, want := ValueA(), "mock-a"; got != want {
		t.Fatalf("after mock ValueA() = %q, want %q", got, want)
	}
	if got, want := ValueB(), "mock-b"; got != want {
		t.Fatalf("after mock ValueB() = %q, want %q", got, want)
	}
}
