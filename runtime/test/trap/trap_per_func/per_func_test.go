package trap_per_func

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trap"
)

// skipsTrap reports whether calling f allocates nothing: a guarded
// function only allocates the arguments of the trap when it calls it, so f
// must not allocate by itself.
func skipsTrap(f func()) bool {
	return testing.AllocsPerRun(100, f) == 0
}

func TestNothingRegisteredSkipsTrap(t *testing.T) {
	if !skipsTrap(func() { Untouched(1) }) {
		t.Fatalf("Untouched calls the trap with nothing registered")
	}
}

func TestPatchEnablesOnlyThatFunction(t *testing.T) {
	mock.Patch(Price, func(cents int64) (int64, error) { return 1, nil })
	if got, _ := Price(5); got != 1 {
		t.Fatalf("Price not mocked: %d", got)
	}
	if !skipsTrap(func() { Tax(5) }) {
		t.Errorf("Tax calls the trap after only Price was patched")
	}
}

func TestPatchBoundMethodEnablesOnlyThatMethod(t *testing.T) {
	c := &Client{prefix: "id-"}
	other := &Client{prefix: "other-"}
	// a bound method value is resolved by calling it with the trap on
	mock.Patch(c.Fetch, func(id int) (string, error) { return "mocked", nil })
	if got, _ := c.Fetch(1); got != "mocked" {
		t.Fatalf("Fetch not mocked: %q", got)
	}
	if got, _ := other.Fetch(1); got != "other-1" {
		t.Fatalf("other instance mocked: %q", got)
	}
	if !skipsTrap(func() { c.Rank(1) }) {
		t.Errorf("Rank calls the trap after only Fetch was patched")
	}
}

func TestFuncInterceptorEnablesOnlyThatFunction(t *testing.T) {
	var seen int
	t.Cleanup(trap.AddFuncInterceptor(Tax, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			seen++
			return nil, nil
		},
	}))
	if got, _ := Tax(50); got != 5 || seen != 1 {
		t.Fatalf("Tax: %d, interceptor calls: %d", got, seen)
	}
	if !skipsTrap(func() { Untouched(1) }) {
		t.Errorf("Untouched calls the trap after only Tax was intercepted")
	}
}

// variables are trapped on every access, skip or not, so mocking one needs
// no flag and must not enable functions
func init() {
	mock.Patch(&InitLimit, func() int { return 30 })
}

func TestPatchVariable(t *testing.T) {
	mock.Patch(&Limit, func() int { return 20 })
	if limit, initLimit := ReadLimits(); limit != 20 || initLimit != 30 {
		t.Fatalf("variables not mocked: Limit=%d InitLimit=%d", limit, initLimit)
	}
	if !skipsTrap(func() { Untouched(1) }) {
		t.Errorf("Untouched calls the trap after a variable was patched")
	}
}

// must stay last: a general interceptor turns the trap on everywhere
func TestZGeneralInterceptorEnablesEverything(t *testing.T) {
	var seen int
	t.Cleanup(trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			if f.IdentityName == "Untouched" {
				seen++
			}
			return nil, nil
		},
	}))
	Untouched(1)
	if seen != 1 {
		t.Fatalf("general interceptor saw Untouched %d times", seen)
	}
	// calls through the trap allocate, which the tests above rely on
	if skipsTrap(func() { Untouched(1) }) {
		t.Errorf("a call through the trap did not allocate")
	}
}
