package mock_var

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

type Config struct {
	On bool
}

var cfg1 Config

const __xgo_trap_cfg2 = 1

var cfg2 Config

func TestConfigAllowIntercept(t *testing.T) {
	var buf bytes.Buffer
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			buf.WriteString(fmt.Sprintf("%s\n", f.IdentityName))
			return
		},
	})
	readConfig1()
	readConfig2()

	trapStr := buf.String()
	// notice there is no lock
	expectTrapStr := "readConfig1\nreadConfig2\ncfg2\n"
	if trapStr != expectTrapStr {
		t.Fatalf("expect tarp buf: %q, actual: %q", expectTrapStr, trapStr)
	}
}

func readConfig1() bool {
	return cfg1.On
}

func readConfig2() bool {
	return cfg2.On
}
