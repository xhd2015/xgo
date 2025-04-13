package global_interceptor_only

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

var trapBuf bytes.Buffer
var testStarted bool

func init() {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if !testStarted {
				// There are some functions callbed before running test
				return nil, nil
			}

			fmt.Fprintf(&trapBuf, "call global %s\n", f.IdentityName)
			return nil, nil
		},
	})
}

func TestGlobalInterceptorOnly(t *testing.T) {
	testStarted = true
	a := A()
	expectA := "A:B"
	if a != expectA {
		t.Fatalf("expect a: %q, actual: %q", expectA, a)
	}
	str := strings.TrimSpace(trapBuf.String())
	expectList := []string{
		"call global A",
		"call global B",
		"call global trapBuf",
	}
	expectStr := strings.Join(expectList, "\n")
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
