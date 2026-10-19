package cross_file

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestMockVarDeclaredInLaterFile(t *testing.T) {
	mock.Patch(&timeout, func() time.Duration { return 3 * time.Second })
	if got := TimeoutSeconds(); got != 3 {
		t.Fatalf("timeout not mocked: %v", got)
	}
}

func TestTrapVarDeclaredInLaterFile(t *testing.T) {
	var names []string
	cancel := trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			names = append(names, f.IdentityName)
			return nil, nil
		},
	})
	Incr()
	cancel()
	if got, want := strings.Join(names, ","), "Incr,mu,mu,count"; got != want {
		t.Fatalf("trapped %q, want %q", got, want)
	}
}
