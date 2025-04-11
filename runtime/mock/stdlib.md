# Automatically instrumented functions that are included in trace
## `os`
- `Getenv`
- `Getwd`
- `OpenFile`

## `time`
- `Now`
- `Sleep`
- `NewTicker`
- `Time.Format`

## `os/exec`
- `Command`
- `(*Cmd).Run`
- `(*Cmd).Output`
- `(*Cmd).Start`

# `net/http`
- `Get`
- `Head`
- `Post`
- `Serve`
- `(*Server).Close`
- `Handle`

# `net`
- `Dial`
- `DialIP`
- `DialUDP`
- `DialUnix`
- `DialTimeout`


# Examples
> Check [../test/mock_stdlib/mock_stdlib_test.go](../test/mock_stdlib/mock_stdlib_test.go) for more details.
```go
package mock_stdlib

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

	// 37.275µs
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
```

Run:`xgo test -v ./`
Output:
```sh
=== RUN   TestMockTimeSleep
--- PASS: TestMockTimeSleep (0.00s)
PASS
ok      github.com/xhd2015/xgo/runtime/test/mock_stdlib 0.725s
```

Note we call `time.Sleep` with `1s`, but it returns within few micro-seconds.