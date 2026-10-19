package instrumented

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"example.com/xgo-initprobe"
	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trap"
	"github.com/xhd2015/xgo/runtime/xgotest"
)

const pkgPath = "github.com/xhd2015/xgo/runtime/test/core/instrumented"

func value() int { return 1 }

func init() {
	if !core.Instrumented() || !initprobe.Instrumented {
		panic("missing xgo linkage in init")
	}
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(context.Context, *core.FuncInfo, core.Object, core.Object) (interface{}, error) {
			return nil, nil
		},
	})()
	mock.Patch(value, func() int { return 2 })
	if value() != 2 {
		panic("patch in init did not take effect")
	}
}

func TestInstrumented(t *testing.T) {
	if !core.Instrumented() {
		t.Fatal("missing xgo linkage")
	}
	for _, mode := range []string{"", "skip", "fail", "invalid"} {
		t.Setenv("XGO_REQUIRE", mode)
		xgotest.Require(t)
	}
	if value() != 2 {
		t.Fatal("init patch lost")
	}
	cancel := mock.PatchByName(pkgPath, "value", func() int { return 3 })
	if value() != 3 {
		t.Fatal("PatchByName failed")
	}
	cancel()
	cancel = mock.Mock(value, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set(4)
		return nil
	})
	if value() != 4 {
		t.Fatal("Mock failed")
	}
	cancel()
	cancel = trap.AddFuncInterceptor(value, &trap.Interceptor{
		Pre: func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) (interface{}, error) {
			results.GetFieldIndex(0).Set(5)
			return nil, trap.ErrMocked
		},
	})
	if value() != 5 {
		t.Fatal("AddFuncInterceptor failed")
	}
	cancel()
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(context.Context, *core.FuncInfo, core.Object, core.Object) (interface{}, error) {
			return nil, nil
		},
	})()
}

func TestUninstrumentedTarget(t *testing.T) {
	for _, tt := range []struct {
		name         string
		fn, replacer interface{}
	}{
		{"stdlib", runtime.GOMAXPROCS, func(int) int { return 1 }},
		{"third-party", initprobe.Untrapped, func() {}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				msg := fmt.Sprint(recover())
				if !strings.Contains(msg, "not instrumented by xgo") || strings.Contains(msg, "binary was not built") {
					t.Fatalf("expected target-specific instrumentation error, got %s", msg)
				}
			}()
			// Hide targets from static mock-ref discovery, which deliberately
			// instruments directly referenced patch targets even in dependencies.
			mock.Patch(tt.fn, tt.replacer)
		})
	}
}
