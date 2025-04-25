package debug

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestDebug(t *testing.T) {
	mock.Patch(greet, func(s string) string {
		return "mock " + s
	})

	got := greet("world")
	want := "mock world"
	if got != want {
		t.Errorf("greet() = %q, want %q", got, want)
	}
}

func greet(s string) string {
	return "hello " + s
}
