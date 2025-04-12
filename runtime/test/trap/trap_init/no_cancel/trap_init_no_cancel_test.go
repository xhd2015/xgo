package trap_init_no_cancel

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trap"
)

var trapBuf bytes.Buffer

func init() {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			trapBuf.WriteString(fmt.Sprintf("call global %s\n", f.IdentityName))
			return
		},
	})
}

func TestTrapCombinedNoCancel(t *testing.T) {
	funcInfo := functab.InfoVar(&trapBuf)
	if funcInfo == nil {
		t.Fatalf("expect trapBuf has funcInfo, actual nil")
	}
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f == funcInfo {
				return nil, nil
			}
			trapBuf.WriteString(fmt.Sprintf("call inside %s\n", f.IdentityName))
			return
		},
	})

	a := A()
	expectA := "A:B"
	if a != expectA {
		t.Fatalf("expect a: %q, actual: %q", expectA, a)
	}
	str := trapBuf.String()
	expectStr := "call global TestTrapCombinedNoCancel\ncall global trapBuf\ncall inside A\ncall global A\ncall inside B\ncall global B\ncall global trapBuf\n"
	if str != expectStr {
		t.Fatalf("expect trap buf: %q, actual: %q", expectStr, str)
	}
}

func A() string {
	return "A:" + B()
}
func B() string {
	return "B"
}
