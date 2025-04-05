package trace_marshal_with_trace

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
)

func TestSubTestTraces(t *testing.T) {
	const N = 10
	traces := make([]*trace.Root, N)
	for i := 0; i < 10; i++ {
		i := i
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			finish := trace.Options().OnComplete(func(root *trace.Root) {
				traces[i] = root
			}).Begin()
			defer finish()
			res := foo(i)

			expect := fmt.Sprintf("foo bar_%d", i)
			if res != expect {
				t.Fatalf("expect %q, actual: %q", expect, res)
			}
		})
	}
	// each should have children
	for i, tr := range traces {
		if len(tr.Children) != 1 {
			t.Fatalf("expect traces[%d] children len to be %d, actual: %d", i, 1, len(tr.Children))
		}
	}
}

func foo(i int) string {
	return "foo " + bar(i)
}
func bar(i int) string {
	return fmt.Sprintf("bar_%d", i)
}
