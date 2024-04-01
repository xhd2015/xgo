package patch

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func greet(s string) string {
	return "hello " + s
}

func greetVaradic(s ...string) string {
	return "hello " + strings.Join(s, ",")
}

func TestPatchSimpleFunc(t *testing.T) {
	mock.Patch(greet, func(s string) string {
		return "mock " + s
	})

	res := greet("world")
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}

func TestPatchVaradicFunc(t *testing.T) {
	mock.Patch(greetVaradic, func(s ...string) string {
		return "mock " + strings.Join(s, ",")
	})

	res := greetVaradic("earth", "moon")
	if res != "mock earth,moon" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock earth,moon", res)
	}
}

type struct_ struct {
	s string
}

func (c *struct_) greet() string {
	return "hello " + c.s
}

func TestPatchMethod(t *testing.T) {
	ins := &struct_{
		s: "world",
	}
	mock.Patch(ins.greet, func() string {
		return "mock " + ins.s
	})

	res := ins.greet()
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}
