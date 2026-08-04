package main

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// Overlay maps helper.go only. --strace may rewrite main.go. Neither should
// clobber the other: helper keeps replacement content and is mockable.
func TestStracePrepAndSiblingOverlay(t *testing.T) {
	if got, want := HelperValue(), "helper-replacement"; got != want {
		t.Fatalf("HelperValue() = %q, want %q", got, want)
	}
	if got, want := MainValue(), "main-original"; got != want {
		t.Fatalf("MainValue() = %q, want %q (main should not be overlaid)", got, want)
	}

	mock.Patch(HelperValue, func() string { return "helper-mocked" })
	if got, want := HelperValue(), "helper-mocked"; got != want {
		t.Fatalf("after mock HelperValue() = %q, want %q", got, want)
	}
	mock.Patch(MainValue, func() string { return "main-mocked" })
	if got, want := MainValue(), "main-mocked"; got != want {
		t.Fatalf("after mock MainValue() = %q, want %q (main should still instrument under --strace)", got, want)
	}
}
