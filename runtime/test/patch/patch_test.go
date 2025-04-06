package patch

import (
	"context"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

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

func TestPatchMethodShouldMatchSameReceiver(t *testing.T) {
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

func TestPatchMethodShouldNotMatchDifferentReceiver(t *testing.T) {
	ins := &struct_{
		s: "world",
	}
	mock.Patch(ins.greet, func() string {
		return "mock " + ins.s
	})

	ins2 := &struct_{
		s: "world",
	}
	res := ins2.greet()
	if res != "hello world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "hello world", res)
	}
}

func TestPatchTypeMethod(t *testing.T) {
	mock.Patch((*struct_).greet, func(s *struct_) string {
		return "mock " + s.s
	})
	ins := &struct_{
		s: "world",
	}
	res := ins.greet()
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}

func TestPatchNilArg(t *testing.T) {
	var haveCalledMock bool
	var argCtx context.Context
	mock.Patch(nilCtx, func(a int, ctx context.Context) {
		argCtx = ctx
		haveCalledMock = true
	})
	nilCtx(0, nil)
	if !haveCalledMock {
		t.Fatalf("expect have called mock,actually not")
	}
	if argCtx != nil {
		t.Fatalf("expect arg ctx to be nil, actual: %v", argCtx)
	}
}
