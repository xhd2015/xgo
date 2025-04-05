package trap_args

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
)

var gc = func(ctx context.Context) {
	panic("gc should be trapped")
}

var gcUnnamed = func(context.Context) {
	panic("gcUnnamed should be trapped")
}

func TestClosureShouldRetrieveCtxInfoAtTrapTime(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "test", "mock")

	// avoid local variable affects interceptor
	gclocal := gc
	callAndCheck(func() {
		gclocal(ctx)
	}, func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) error {
		if !f.FirstArgCtx {
			t.Fatalf("expect closure also mark firstArgCtx, actually not marked")
		}
		if trapCtx == nil {
			t.Fatalf("expect trapCtx to be non nil, atcual nil")
		}
		if trapCtx != ctx {
			t.Fatalf("expect trapCtx to be the same with ctx, actually different")
		}
		return nil
	})
}

func TestClosureUnnamedArgShouldRetrieveCtxInfo(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "test", "mock")
	localGcUnnamed := gcUnnamed
	callAndCheck(func() {
		localGcUnnamed(ctx)
	}, func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) error {
		if !f.FirstArgCtx {
			t.Fatalf("expect closure also mark firstArgCtx, actually not marked")
		}
		if trapCtx == nil {
			t.Fatalf("expect trapCtx to be non nil, atcual nil")
		}
		if trapCtx != ctx {
			t.Fatalf("expect trapCtx to be the same with ctx, actually different")
		}
		return nil
	})
}
