package trap

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestTrapAbortInTheMiddle(t *testing.T) {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			if f.IdentityName == "double" {
				panic("should be aborted")
			}
			return nil, nil
		},
	})
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			if f.IdentityName == "double" {
				args.GetField("i").Set(20)
				return nil, trap.ErrAbort
			}
			return nil, nil
		},
	})

	d := double(1)
	if d != 0 {
		t.Fatalf("expect double(1) to be aborted so return %d actual: %d", 0, d)
	}
}

func double(i int) int {
	return 2 * i
}
