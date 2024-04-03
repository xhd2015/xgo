// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"context"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func TestMockTimeSleep(t *testing.T) {
	begin := time.Now()
	sleepDur := 1 * time.Second
	var haveCalledMock bool
	var mockArg time.Duration
	mock.Mock(time.Sleep, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		haveCalledMock = true
		mockArg = args.GetFieldIndex(0).Value().(time.Duration)
		return nil
	})
	time.Sleep(sleepDur)

	cost := time.Since(begin)

	if !haveCalledMock {
		t.Fatalf("expect haveCalledMock, actually not")
	}
	if mockArg != sleepDur {
		t.Fatalf("expect mockArg to be %v, actual: %v", sleepDur, mockArg)
	}
	if cost > 100*time.Millisecond {
		t.Fatalf("expect time.Sleep mocked, actual cost: %v", cost)
	}
}
