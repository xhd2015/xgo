package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func main() {
	if os.Getenv("XGO_TEST_HAS_INSTRUMENT") != "false" {
		mock.AddFuncInterceptor(hello, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
			a := args.GetField("a")
			a.Set("mock:" + a.Value().(string))
			return mock.ErrCallOld
		})
	}
	hello("world")
}

func hello(a string) {
	fmt.Printf("hello %s\n", a)
}
