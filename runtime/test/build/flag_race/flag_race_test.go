package flag_race

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

// bug issue: https://github.com/xhd2015/xgo/issues/330
func TestFlagRace(t *testing.T) {
	var buf bytes.Buffer
	trap.AddFuncInterceptor(os.Getenv, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			buf.WriteString(f.IdentityName + "\n")
			return nil, nil
		},
	})

	os.Getenv("GORACE")

	res := buf.String()
	if res != "Getenv\n" {
		t.Fatalf("expect Getenv, but got %s", res)
	}
}
