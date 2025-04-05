package main

import (
	"context"
	"fmt"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func main() {
	before := add(5, 2)
	fmt.Printf("before mock: add(5,2)=%d\n", before)
	mock.Mock(add, func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error {
		a := args.GetField("a").Value().(int)
		b := args.GetField("b").Value().(int)
		res := a - b
		results.GetFieldIndex(0).Set(res)
		return nil
	})
	after := add(5, 2)
	fmt.Printf("after mock: add(5,2)=%d\n", after)
}

func add(a int, b int) int {
	return a + b
}
