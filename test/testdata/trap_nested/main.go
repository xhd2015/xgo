package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

// Prior to xgo v1.0.1, explicit trap.Skip()
// must be made to avoid stack overflow.
// This is called 'the nested trap' problem.
//
// After xgo v1.0.1,xgo can check dynamically
// whether current goroutine is inside trap, if so
// skip it

func main() {
	if os.Getenv("XGO_TEST_HAS_INSTRUMENT") != "false" {
		trap.AddInterceptor(&trap.Interceptor{
			Pre: func(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object) (interface{}, error) {
				fmt.Printf("trap pre: %s\n", f.Name)
				callFromTrap()
				return nil, nil
			},
		})
	}
	hello()
}

func hello() {
	fmt.Printf("hello world\n")
}

func callFromTrap() {
	fmt.Printf("call from trap\n")
}
