package main

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/xhd2015/xgo/test/example/closure/sub"
)

func main() {
	var a = func() {}

	// main.main.func1
	fnNameA := runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	fmt.Printf("a: %s\n", fnNameA)

	// github.com/xhd2015/xgo/test/example/closure/sub.glob..func1
	fnName := runtime.FuncForPC(reflect.ValueOf(sub.F).Pointer()).Name()
	fmt.Printf("sub.F: %s\n", fnName)

	// github.com/xhd2015/xgo/test/example/closure/sub.init.0.func1
	fnNameZ := runtime.FuncForPC(reflect.ValueOf(sub.Z).Pointer()).Name()
	fmt.Printf("sub.Z: %s\n", fnNameZ)
}
