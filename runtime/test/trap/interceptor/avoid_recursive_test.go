package interceptor

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

// prints: pre->call_f->post
// no repeation
func TestNakedTrapShouldAvoidRecursiveInterceptor(t *testing.T) {
	var recurseBuf bytes.Buffer
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Stdlib {
				return nil, nil
			}
			fmt.Fprintf(&recurseBuf, "pre\n")
			return nil, nil
		},
		Post: func(ctx context.Context, f *core.FuncInfo, args, result core.Object, data interface{}) (err error) {
			if f.Stdlib {
				return nil
			}
			fmt.Fprintf(&recurseBuf, "post\n")
			return nil
		},
	})
	f(&recurseBuf)
	output := recurseBuf.String()
	expect := "pre\ncall_f\npost\n"
	if output != expect {
		t.Fatalf("expect no recursive trap, output to be %q, actual: %q", expect, output)
	}
}
