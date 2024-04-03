package trap

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestDirectShouldByPassTrap(t *testing.T) {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "direct" {
				panic("direct should be bypassed")
			}
			if f.IdentityName == "nonByPass" {
				result.GetFieldIndex(0).Set("mock nonByPass")
				return nil, trap.ErrAbort
			}
			return nil, nil
		},
	})
	var resDirect string
	trap.Direct(func() {
		resDirect = direct()
	})
	if resDirect != "direct" {
		t.Fatalf("expect direct() to be %q, actual: %q", "direct", resDirect)
	}

	nonRes := nonByPass()
	if nonRes != "mock nonByPass" {
		t.Fatalf("expect nonByPass() to be %q, actual: %q", "mock nonByPass", nonRes)
	}
}

func direct() string {
	return "direct"
}
func nonByPass() string {
	return "nonByPass"
}
