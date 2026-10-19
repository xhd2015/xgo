//go:build go1.18
// +build go1.18

package trap_guard

import (
	"testing"
)

func TestUnmockedGeneric(t *testing.T) {
	if v, ok := First(3, 4); !ok || v != 3 {
		t.Errorf("First: %v %v", v, ok)
	}
	if _, ok := First[int](); ok {
		t.Errorf("First of nothing")
	}
}
