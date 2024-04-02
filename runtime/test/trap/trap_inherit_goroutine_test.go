package trap

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestNewGoroutineShouldInheritInterceptor(t *testing.T) {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "greet" {
				result.GetFieldIndex(0).Set("mock " + args.GetFieldIndex(0).Value().(string))
				return nil, trap.ErrAbort
			}
			return
		},
	})

	done := make(chan struct{})
	go func() {
		s := greet("world")
		if s != "mock world" {
			panic(fmt.Errorf("expect greet returns %q, actual: %q", "mock world", s))
		}

		close(done)
	}()

	<-done
}

func greet(s string) string {
	return "hello " + s
}
