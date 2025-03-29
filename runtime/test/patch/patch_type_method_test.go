package patch

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func TestPatchTypeMethodNoArg(t *testing.T) {
	ins := &struct_{
		s: "world",
	}
	mock.Patch((*struct_).greet, func(ins *struct_) string {
		return "mock " + ins.s
	})

	res := ins.greet()
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}

func TestPatchTypeMethodCtxArg(t *testing.T) {
	ins := &struct_{
		s: "world",
	}
	mock.Patch((*struct_).greetCtx, func(ins *struct_, ctx context.Context) string {
		return "mock " + ins.s
	})

	res := ins.greetCtx(context.Background())
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}

func TestPatchInstanceMethodCtxArg(t *testing.T) {
	ins := &struct_{
		s: "world",
	}
	mock.Patch(ins.greetCtx, func(ctx context.Context) string {
		return "mock " + ins.s
	})

	res := ins.greetCtx(context.Background())
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}
