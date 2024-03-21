package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/trace"
)

func init() {
	trace.Enable()
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("main panic: %v", e)
		}
	}()
	fmt.Printf("call: main\n")
	Work()
}
func Work() {
	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("Work panic: %v", e))
		}
	}()
	fmt.Printf("call: Work\n")
	doWork()
}

func doWork() {
	fmt.Printf("call: doWork\n")
	panic("doWork panic")
}
