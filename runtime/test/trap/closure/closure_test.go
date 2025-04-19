package rule

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func getClosure() func() {
	return func() {

	}
}
func TestClosureDefaultMock(t *testing.T) {
	t.Skip("closure is not supported since xgo v1.1.0")
	testClosureDefaultMock(t, "call getClosure\ncall getClosure.func1\n")
}

// flags: --mock-rule '{"closure":true,"action":"exclude"}'
func TestClosureWithMockRuleNoMock(t *testing.T) {
	t.Skip("closure is not supported since xgo v1.1.0")
	testClosureDefaultMock(t, "call getClosure\n")
}

func testClosureDefaultMock(t *testing.T, expect string) {
	t.Helper()
	var buf strings.Builder
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return
			}
			buf.WriteString(fmt.Sprintf("call %s\n", f.IdentityName))
			return
		},
	})
	closure := getClosure()
	closure()

	got := buf.String()
	if got != expect {
		t.Fatalf("interceptor expect: %q, actual: %q", expect, got)
	}
}
