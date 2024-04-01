package trap_args

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
)

var gc = func(ctx context.Context) {
	panic("gc should be trapped")
}

func TestClosureShouldRetrieveCtxInfoAtTrapTime(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "test", "mock")
	callAndCheck(func() {
		gc(ctx)
	}, func(trapCtx context.Context, f *core.FuncInfo, args, result core.Object) error {
		if !f.FirstArgCtx {
			t.Fatalf("expect closure also mark firstArgCtx, actually not marked")
		}
		if trapCtx == nil {
			t.Fatalf("expect trapCtx to be non nil, atcual nil")
		}
		if trapCtx != ctx {
			t.Fatalf("expect trapCtx to be the same with ctx, actully different")
		}
		return nil
	})
}
