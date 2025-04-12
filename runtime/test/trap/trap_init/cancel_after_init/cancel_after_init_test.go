package cancel_after_init

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

var cancel func()

func init() {
	cancel = trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			return
		},
	})
}

func TestTrapCombinedNoCancel(t *testing.T) {
	var panicErr interface{}
	func() {
		defer func() {
			panicErr = recover()
		}()
		cancel()
	}()
	if panicErr == nil {
		t.Fatalf("expect panic, actual no panic")
	}
	panicMsg := fmt.Sprint(panicErr)
	expectMsg := "global interceptor cannot be cancelled after init finished"
	if panicMsg != expectMsg {
		t.Fatalf("expect panic msg: %q, actual: %q", expectMsg, panicMsg)
	}
}
