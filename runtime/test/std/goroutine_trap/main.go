package main

import (
	"context"
	"fmt"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

var done chan struct{}

func main() {
	done = make(chan struct{})

	go printHello()
	<-done

	fmt.Printf("main: %s\n", goroutineStr())
}

func printHello() {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			fmt.Printf("call: %s\n", f.DisplayName())
			return nil, nil
		},
	})
	fmt.Printf("printHello: %s\n", goroutineStr())
	close(done)
}

func goroutineStr() string {
	return "goroutine"
}
