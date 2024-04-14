package trap_init

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

var trapBuf bytes.Buffer
var initA string

func init() {
	cancel := trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			trapBuf.WriteString(fmt.Sprintf("call %s\n", f.IdentityName))
			if f.IdentityName == "B" {
				result.GetFieldIndex(0).Set("mock B")
				return nil, trap.ErrAbort
			}
			return
		},
	})
	defer cancel()
	initA = A()
}

func TestTrapInsideInit(t *testing.T) {
	str := trapBuf.String()
	expectTrapBuf := "call A\ncall B\n"
	if str != expectTrapBuf {
		t.Fatalf("expect trap buf: %q, actual: %q", expectTrapBuf, str)
	}
	expectInitA := "A:mock B"
	if initA != expectInitA {
		t.Fatalf("expect initA: %q, actual: %q", expectInitA, initA)
	}

	// check if the interceptor is cancelled
	a := A()
	expectA := "A:B"
	if a != expectA {
		t.Fatalf("expect a: %q, actual: %q", expectA, a)
	}
}

func A() string {
	return "A:" + B()
}
func B() string {
	return "B"
}
