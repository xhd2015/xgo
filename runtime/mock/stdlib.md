# Support Of Mocking Stdlib Functions
There are two modes for mocking stdlib functions,
1. Default-Allow Mode
2. Default-Disallow Mode

These two modes can be switched by `--trap-stdlib` and `--trap-stdlib=false`.

When running `xgo test`, `--trap-stdlib` is assumed, you can turn it off with `--trap-stdlib=false`.

When running `xgo run` and `xgo build`, `--trap-stdlib=false` is assumed, you can turn it on with `--trap-stdlib`.

# `--trap-stdlib`: Default-Allow Mode
In this mode, most stdlib packages can be mocked, except `runtime`, `syscall`, `reflect`, `sync`, `sync/atomic`, `testing`, `unsafe`.

# `--trap-stdlib=false`: Default-Disallow Mode
In this mode, only a small list of stdlib functions can be mocked due to performance and security impact.

Rational: stdlib functions like `os.ReadFile`, `io.Read` are widely used by go code, installing a trap on these functions may have big impact on performance and security.

So in this mode only a limited list of stdlib functions can be mocked. However, if there lacks some functions you may want to use, you can leave a comment in [Issue#6](https://github.com/xhd2015/xgo/issues/6) or fire an issue to let us know and add it.

# Functions In Stdlib Calling `recover()` Cannot Be Mocked
When a function calls `recover()`, it will capture panic when used in defer.

However, since compiler treats stdlib from ordinary module differently, current implementation to support stdlib function is based on source code injection, which may causes build time to slow down, and also causes functions containing `recover()` to be invalid if rewritten, see https://github.com/xhd2015/xgo/issues/164.

This will be fixed in the long run, but before that, such functions in stdlib cannot be mocked.

NOTE: functions outside stdlib, even with calling `recover()`, are not affected since they are rewritten with IR, not source code.

# Supported List When `--trap-stdlib=false`
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