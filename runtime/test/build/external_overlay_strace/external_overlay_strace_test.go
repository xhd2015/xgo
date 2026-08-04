package main

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// --strace prep and -overlay must compose: caller replacement body is built,
// instrumented, and mockable (prep must not wipe the overlay with a copy of
// the original main.go).
func TestExternalOverlayWithStraceKeepsReplacementAndInstruments(t *testing.T) {
	if got, want := Value(), "caller replacement"; got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
	mock.Patch(Value, func() string { return "mocked" })
	if got, want := Value(), "mocked"; got != want {
		t.Fatalf("after mock Value() = %q, want %q", got, want)
	}
}
