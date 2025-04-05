package trap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func f(recurseBuf io.Writer) {
	fmt.Fprintf(recurseBuf, "call_f\n")
}

func TestDeferredFuncShouldBeExecutedWhenAbort(t *testing.T) {
	var recurseBuf bytes.Buffer
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			fmt.Fprintf(&recurseBuf, "pre\n")
			return nil, trap.ErrAbort
		},
		Post: func(ctx context.Context, f *core.FuncInfo, args, result core.Object, data interface{}) (err error) {
			if f.Stdlib {
				return nil
			}
			fmt.Fprintf(&recurseBuf, "post\n")
			return nil
		},
	})
	// NOTE: f's body is skipped
	f(&recurseBuf)
	output := recurseBuf.String()
	expect := "pre\npost\n"
	if output != expect {
		t.Fatalf("expect no recursive trap, output to be %q, actual: %q", expect, output)
	}
}
