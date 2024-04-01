package trap_args

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
)

func TestPlainCtxArgCanBeRecognized(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "test", "mock")
	callAndCheck(func() {
		f2(ctx)
	}, func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) error {
		if !f.FirstArgCtx {
			t.Fatalf("expect first arg to be context")
		}
		if trapCtx != ctx {
			t.Fatalf("expect context passed unchanged, actually different")
		}
		ctxVal := trapCtx.Value("test").(string)
		if ctxVal != "mock" {
			t.Fatalf("expect context value to be %q, actual: %q", "mock", ctxVal)
		}
		return nil
	})
}

func TestCtxVariantCanBeRecognized(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "test", "mock")

	myCtx := &MyContext{Context: ctx}

	callAndCheck(func() {
		f3(myCtx)
	}, func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) error {
		if !f.FirstArgCtx {
			t.Fatalf("expect first arg to be context")
		}
		if trapCtx != myCtx {
			t.Fatalf("expect context passed unchanged, actually different")
		}
		ctxVal := trapCtx.Value("test").(string)
		if ctxVal != "mock" {
			t.Fatalf("expect context value to be %q, actual: %q", "mock", ctxVal)
		}
		return nil
	})
}
