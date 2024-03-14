package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xhd2015/xgo/runtime/core/trap"
)

func init() {
	if os.Getenv("XGO_TEST_NO_INSTRUMENT") != "true" {
		trap.AddInterceptor(&trap.Interceptor{
			Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
				trap.Skip()
				fmt.Printf("global trap: %s\n", f.Name)
				return nil, nil
			},
		})
	}
}

func main() {
	A()
	B()
}

func A() {
	// non instrument mode cannot get curg
	if os.Getenv("XGO_TEST_NO_INSTRUMENT") != "true" {
		cancel := trap.AddLocalInterceptor(&trap.Interceptor{
			Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
				trap.Skip()
				fmt.Printf("local trap from A: %s\n", f.Name)
				return nil, nil
			},
		})
		defer cancel()
	}
	printHello()
}

func B() {
	if os.Getenv("XGO_TEST_NO_INSTRUMENT") != "true" {
		cancel := trap.AddLocalInterceptor(&trap.Interceptor{
			Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
				trap.Skip()
				fmt.Printf("local trap from B: %s\n", f.Name)
				return nil, nil
			},
		})
		defer cancel()
	}
	printHello()
}

func printHello() {
	fmt.Printf("hello\n")
}
