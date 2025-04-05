package interceptor

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestNewGoroutineShouldInheritInterceptor(t *testing.T) {
	var recordGreet string
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "greet" {
				recordGreet = "recorded: " + args.GetFieldIndex(0).Value().(string)
			}
			return
		},
	})

	done := make(chan struct{})
	go func() {
		s := greet("world")
		if s != "hello world" {
			panic(fmt.Errorf("expect greet returns %q, actual: %q", "hello world", s))
		}

		close(done)
	}()

	<-done
	if recordGreet != "recorded: world" {
		panic(fmt.Errorf("expect recordGreet to be %q, actual: %q", "recorded: world", recordGreet))
	}
}

func greet(s string) string {
	return "hello " + s
}
