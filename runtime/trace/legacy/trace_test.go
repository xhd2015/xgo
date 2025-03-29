package legacy

import (
	"context"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestNowOriginal(t *testing.T) {
	var captured bool
	trap.AddFuncInterceptor(time.Now, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Pkg == "time" && f.IdentityName == "Now" {
				captured = true
			}
			return nil, nil
		},
	})
	time.Now()
	if !captured {
		t.Fatalf("should be captured")
	}
}

func TestNowPatched(t *testing.T) {
	var captured bool
	trap.AddFuncInterceptor(time.Now, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Pkg == "time" && f.IdentityName == "Now" {
				captured = true
			}
			return nil, nil
		},
	})
	timeNow()
	if captured {
		t.Fatalf("should not be captured")
	}
}
