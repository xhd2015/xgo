package mock_var

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	internal_trap "github.com/xhd2015/xgo/runtime/internal/trap"
	"github.com/xhd2015/xgo/runtime/trap"
)

type Config struct {
	On bool
}

var cfg1 Config

const __xgo_trap_cfg2 = 1

var cfg2 Config

func TestConfigAllowIntercept(t *testing.T) {
	trap.MarkIntercept((*bytes.Buffer).String)
	var buf bytes.Buffer
	_, bufFn, _, _ := internal_trap.Inspect((*bytes.Buffer).String)
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f == bufFn {
				return
			}
			buf.WriteString(fmt.Sprintf("call %s\n", f.IdentityName))
			return
		},
	})
	readConfig1()
	readConfig2()
	_ = cfg2

	trapStr := buf.String()
	// notice there is no lock
	expectTrapStr := "call readConfig1\ncall cfg1\ncall readConfig2\ncall cfg2\ncall cfg2\n"
	if trapStr != expectTrapStr {
		t.Fatalf("expect trap buf: %q, actual: %q", expectTrapStr, trapStr)
	}
}

func readConfig1() bool {
	return cfg1.On
}

func readConfig2() bool {
	return cfg2.On
}
