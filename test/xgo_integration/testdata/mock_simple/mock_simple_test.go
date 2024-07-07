package mock_simple

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func greet(s string) string {
	return "hello " + s
}

func TestMockSimple(t *testing.T) {
	mock.Patch(greet, func(s string) string {
		return "mock " + s
	})
	s := greet("world")
	if s != "mock world" {
		t.Fatalf("expect greet(): %q, actual: %q", "mock world", s)
	}
}
