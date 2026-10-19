package trap_guard

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"sync"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestInitGoroutineRunsUnmocked(t *testing.T) {
	<-initGoroutineDone
	if got := initGoroutineResult; got != "charged init 1" {
		t.Fatalf("init goroutine: %q", got)
	}
}

func TestUnmockedCallsKeepBehavior(t *testing.T) {
	if got := CallCharge("a"); got != "charged a 100" {
		t.Errorf("Charge: %q", got)
	}
	s := &Store{prefix: "p:"}
	if v, err := CallGet(s, "k"); err != nil || v != "p:k" {
		t.Errorf("Get: %q %v", v, err)
	}
	if _, err := CallGet(s, ""); err == nil {
		t.Errorf("Get: expected error")
	}
	if got := (Store{prefix: "x"}).Name(); got != "store:x" {
		t.Errorf("Name: %q", got)
	}
	if got := Join(",", "a", "b"); got != "a,b" {
		t.Errorf("Join: %q", got)
	}
	if got := Normalize("  ab "); got != "AB" {
		t.Errorf("Normalize: %q", got)
	}
	if got := Increment(1); got != 2 {
		t.Errorf("Increment: %d", got)
	}
}

func TestPatchFunc(t *testing.T) {
	mock.Patch(Charge, func(ctx context.Context, customer string, cents int64) string {
		return "mocked " + customer
	})
	if got := CallCharge("b"); got != "mocked b" {
		t.Fatalf("Charge: %q", got)
	}
}

func TestPatchMethodExpression(t *testing.T) {
	mock.Patch((*Store).Get, func(s *Store, ctx context.Context, key string) (string, error) {
		return "mocked " + s.prefix + key, nil
	})
	if v, err := CallGet(&Store{prefix: "p:"}, "k"); err != nil || v != "mocked p:k" {
		t.Fatalf("Get: %q %v", v, err)
	}
}

func TestMockMethodByName(t *testing.T) {
	s := &Store{prefix: "p:"}
	other := &Store{prefix: "o:"}
	mock.MockMethodByName(s, "Get", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("by name " + args.GetField("key").Value().(string))
		return nil
	})
	if v, _ := CallGet(s, "k"); v != "by name k" {
		t.Errorf("mocked instance: %q", v)
	}
	if v, _ := CallGet(other, "k"); v != "o:k" {
		t.Errorf("other instance must not be mocked: %q", v)
	}
}

func TestInterceptorErrMocked(t *testing.T) {
	t.Cleanup(trap.AddFuncInterceptor((*Store).Get, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			result.GetFieldIndex(0).Set("intercepted")
			return nil, trap.ErrMocked
		},
	}))
	if v, err := CallGet(&Store{}, ""); err != nil || v != "intercepted" {
		t.Fatalf("Get: %q %v", v, err)
	}
}

// Pre may rewrite arguments; the function body must see the new value.
func TestInterceptorRewritesArgument(t *testing.T) {
	t.Cleanup(trap.AddFuncInterceptor(Increment, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			args.GetField("n").Set(10)
			return nil, nil
		},
	}))
	if got := Increment(1); got != 11 {
		t.Fatalf("Increment: %d", got)
	}
}

// Post observes results after the body and its defers ran, and may rewrite
// them.
func TestInterceptorPostSeesAndRewritesResults(t *testing.T) {
	var seen string
	t.Cleanup(trap.AddFuncInterceptor(Normalize, &trap.Interceptor{
		Post: func(ctx context.Context, f *core.FuncInfo, args, result core.Object, data interface{}) error {
			seen = result.GetField("out").Value().(string)
			result.GetField("out").Set(seen + "!")
			return nil
		},
	}))
	if got := Normalize(" x "); got != "X!" {
		t.Errorf("Normalize: %q", got)
	}
	if seen != "X" {
		t.Errorf("post saw %q", seen)
	}
}

func TestPostRunsOnPanic(t *testing.T) {
	var ran bool
	t.Cleanup(trap.AddFuncInterceptor((*Store).Get, &trap.Interceptor{
		Post: func(ctx context.Context, f *core.FuncInfo, args, result core.Object, data interface{}) error {
			ran = true
			return nil
		},
	}))
	func() {
		defer func() { recover() }()
		var s *Store
		CallGet(s, "k") // nil receiver panics in the body
	}()
	if !ran {
		t.Errorf("post did not run")
	}
}

// Mocks are per goroutine and inherited by goroutines started afterwards.
func TestPatchInheritedByGoroutines(t *testing.T) {
	mock.Patch(Charge, func(ctx context.Context, customer string, cents int64) string {
		return "child " + customer
	})
	var wg sync.WaitGroup
	results := make([]string, 4)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = CallCharge("c")
		}(i)
	}
	wg.Wait()
	for _, r := range results {
		if r != "child c" {
			t.Fatalf("goroutine result: %q", r)
		}
	}
}

func TestParallelPatchesAreIsolated(t *testing.T) {
	for _, name := range []string{"x", "y", "z"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mock.Patch(Charge, func(ctx context.Context, customer string, cents int64) string {
				return name
			})
			for i := 0; i < 100; i++ {
				if got := CallCharge("p"); got != name {
					t.Fatalf("got %q", got)
				}
			}
		})
	}
}

func TestPatchErrorResult(t *testing.T) {
	want := errors.New("boom")
	mock.Patch((*Store).Get, func(s *Store, ctx context.Context, key string) (string, error) {
		return "", want
	})
	if _, err := CallGet(&Store{}, "k"); err != want {
		t.Fatalf("err: %v", err)
	}
}

// A result stored by a pre-hook that does not stop the call is visible to
// the body, as with the result variable itself.
func TestPreResultVisibleToBody(t *testing.T) {
	if got := Lookup("k"); got != "default:k" {
		t.Fatalf("Lookup: %q", got)
	}
	t.Cleanup(trap.AddFuncInterceptor(Lookup, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			result.GetField("value").Set("pre")
			return nil, nil
		},
	}))
	if got := Lookup("k"); got != "pre+body" {
		t.Fatalf("Lookup: %q", got)
	}
}

// The trap prologue must not declare closures of its own: they would shift
// the numbering of the function's closures (Handlers.func1 -> func2).
func TestClosureNamesUnchanged(t *testing.T) {
	a, b := Handlers()
	for _, c := range []struct {
		fn   func() string
		want string
	}{
		{a, "github.com/xhd2015/xgo/runtime/test/trap/trap_guard.Handlers.func1"},
		{b, "github.com/xhd2015/xgo/runtime/test/trap/trap_guard.Handlers.func2"},
	} {
		if got := runtime.FuncForPC(reflect.ValueOf(c.fn).Pointer()).Name(); got != c.want {
			t.Errorf("closure name: %s, want %s", got, c.want)
		}
	}
}
