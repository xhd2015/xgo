package mock_stdlib

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

// go run ./cmd/xgo test --project-dir runtime -run TestMockTimeNow -v ./test/mock_stdlib
// go run ./script/run-test/ --include go1.22.1 --xgo-runtime-test-only -run TestMockTimeNow -v ./test/mock_stdlib
func TestMockTimeNow(t *testing.T) {
	now1 := time.Now()
	now2 := time.Now()

	d1 := now2.Sub(now1)
	if d1 <= 0 {
		t.Fatalf("expect now2-now1 > 0 , actual: %v", d1)
	}
	cancel := mock.Mock(time.Now, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set(now1)
		return nil
	})

	now3 := time.Now()
	if now3 != now1 {
		t.Fatalf("expect now3 equals to now1 exactly, actual diff: %v", now3.Sub(now1))
	}
	cancel()

	now4 := time.Now()
	d4 := now4.Sub(now1)
	if d4 <= 0 {
		t.Fatalf("expect now4-now1 > 0 after cancelling mock, actual: %v", d4)
	}
}

// go run ./cmd/xgo test --project-dir runtime -run TestMockHTTP -v ./test/mock_stdlib
func TestMockHTTP(t *testing.T) {
	var haveMocked bool
	mock.Mock(http.DefaultClient.Do, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		haveMocked = true
		return nil
	})
	http.DefaultClient.Do(nil)
	if !haveMocked {
		t.Fatalf("expect http.DefaultClient.Do to have been mocked, actually not mocked")
	}
}

// execution log(NOTE the cost is not 1s):
//
// === RUN   TestMockTimeSleep
// --- PASS: TestMockTimeSleep (0.00s)
// PASS
// ok      github.com/xhd2015/xgo/runtime/test/mock_stdlib 0.725s
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

	// 37.275µs
	cost := time.Since(begin)

	if !haveCalledMock {
		t.Fatalf("expect haveCalledMock, actually not")
	}
	if mockArg != sleepDur {
		t.Fatalf("expect mockArg to be %v, actual: %v", sleepDur, mockArg)
	}
	// t.Logf("cost: %v", cost)
	if cost > 100*time.Millisecond {
		t.Fatalf("expect time.Sleep mocked, actual cost: %v", cost)
	}
}
