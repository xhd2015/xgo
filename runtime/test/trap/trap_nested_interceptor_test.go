package trap

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestNestedTrapShouldBeAllowedBySpecifyingMapping(t *testing.T) {
	var traceRecords bytes.Buffer

	// list the names that can be nested
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			traceRecords.WriteString(fmt.Sprintf("call %s\n", f.IdentityName))
			if f.IdentityName == "A0" {
				// call A1 inside the interceptor
				result.GetFieldIndex(0).Set(A1())
				return nil, trap.ErrAbort
			}
			return nil, nil
		},
	})
	trap.AddFuncInterceptor(A1, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			traceRecords.WriteString(fmt.Sprintf("call %s\n", f.IdentityName))
			return
		},
	})
	trap.AddFuncInterceptor(A2, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			traceRecords.WriteString(fmt.Sprintf("call %s\n", f.IdentityName))
			return
		},
	})

	result := A0()

	trace := traceRecords.String()
	expect := "call A0\ncall A1\ncall A2\n"
	if trace != expect {
		t.Fatalf("expect trace: %q, actual: %q", expect, trace)
	}
	expectResult := "A1 A2"
	if result != expectResult {
		t.Fatalf("expect result: %q, actual: %q", expectResult, result)
	}
}
func A0() string {
	return "A0"
}

func A1() string {
	return "A1 " + A2()
}

func A2() string {
	return "A2"
}

func TestNestedTrapPartialAllowShouldTakeEffect(t *testing.T) {
	var traceRecords bytes.Buffer
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			traceRecords.WriteString(fmt.Sprintf("call %s\n", f.IdentityName))
			if f.IdentityName == "B0" {
				// call A1 inside the interceptor
				result.GetFieldIndex(0).Set(B1())
				return nil, trap.ErrAbort
			}
			return nil, nil
		},
	})

	trap.AddFuncInterceptor(B2, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			traceRecords.WriteString(fmt.Sprintf("call %s\n", f.IdentityName))
			return
		},
	})
	result := B0()

	trace := traceRecords.String()
	expect := "call B0\ncall B2\n"
	if trace != expect {
		t.Fatalf("expect trace: %q, actual: %q", expect, trace)
	}
	expectResult := "B1 B2"
	if result != expectResult {
		t.Fatalf("expect result: %q, actual: %q", expectResult, result)
	}
}
func B0() string {
	return "B0"
}

func B1() string {
	return "B1 " + B2()
}

// B2 is ignored
func B2() string {
	return "B2"
}

func TestTimeNowNestedLevel1Normal(t *testing.T) {
	var r time.Time
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if r.IsZero() && f == functab.InfoFunc(time.Now) {
				r = time.Now()
			}
			return
		},
	})
	now := time.Now()

	diff := now.Sub(r)
	if diff < 0 {
		t.Fatalf("pre should happen before call, diff: %v", diff)
	}
	// on windows this will exceed 1ms
	threshold := 1 * time.Millisecond
	if runtime.GOOS == "windows" {
		threshold = 5 * time.Millisecond
	}
	if diff > threshold {
		t.Fatalf("interval too large:%v", diff)
	}
}

func TestTimeNowNestedLevel2Normal(t *testing.T) {
	var r time.Time
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if r.IsZero() && f == functab.InfoFunc(time.Now) {
				r = getTime()
			}
			return
		},
	})
	now := time.Now()

	diff := now.Sub(r)
	if diff < 0 {
		t.Fatalf("pre should happen before call, diff: %v", diff)
	}

	// on windows this will exceed 1ms
	threshold := 1 * time.Millisecond
	if runtime.GOOS == "windows" {
		threshold = 5 * time.Millisecond
	}
	if diff > threshold {
		t.Fatalf("interval too large:%v", diff)
	}
}

func getTime() time.Time {
	return time.Now()
}
