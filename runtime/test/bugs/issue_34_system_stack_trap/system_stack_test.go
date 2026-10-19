package system_stack_trap

import (
	"context"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestTimerChannelSystemStack(t *testing.T) {
	calls := 0
	defer trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, results core.Object) (interface{}, error) {
			if f.Pkg == "time" && f.IdentityName == "Now" {
				calls++
			}
			return nil, nil
		},
	})()

	// The interceptor, and the goroutine-local trap state, must be inherited.
	done := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		defer close(done)
		for i := 0; i < 10; i++ {
			timer := time.NewTimer(time.Nanosecond)
			ticker := time.NewTicker(time.Nanosecond)
			// Neither channel has been received from, so neither timer is in
			// the timer heap. Once expired, select runs sendTime on g0 via
			// runtime.(*timer).maybeRunChan rather than the timer goroutine.
			time.Sleep(time.Millisecond)
			select {
			case <-timer.C:
			case <-stop:
				return
			}
			select {
			case <-ticker.C:
			case <-stop:
				return
			}
			ticker.Stop()
			time.Now()
		}
	}()
	<-done
	if calls < 10 {
		t.Fatalf("ordinary goroutine calls must still be intercepted: got %d, want at least 10", calls)
	}
}
