package main

import (
	"context"
	"fmt"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

// go run ./cmd/xgo run --project-dir ./runtime ./test/mock
func main() {
	OnFunc(hello, func(a string) {
		if a == "" {
			CallOld()
			// be able to call old function
		}

		return
	})

	mock.AddFuncInterceptor(hello, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		a := args.GetField("a")
		a.Set("mock:" + a.Value().(string))
		return mock.ErrCallOld
	})
	hello("world")
}

func hello(a string) {
	fmt.Printf("hello %s\n", a)
}

func OnFunc[T any](fn T, v T) {

}

func CallOld() {

}
