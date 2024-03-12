package main

import (
	"context"
	"fmt"

	"github.com/xhd2015/xgo/runtime/core/trap"
)

func init() {
	trap.Use()
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
			trap.Skip()
			if f.Name == "A" {
				fmt.Printf("trap A\n")
				return nil, nil
			}
			if f.Name == "B" {
				fmt.Printf("abort B\n")
				return nil, trap.ErrAbort
			}
			return nil, nil
		},
	})
}

func main() {
	A()
	B()
}

func A() {
	fmt.Printf("A\n")
}

func B() {
	fmt.Printf("B\n")
}
