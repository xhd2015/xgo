package main

import (
	"fmt"
	"reflect"
	"runtime"
)

type SomeType struct {
}

func (s *SomeType) SomeFunc() {

}

func main() {
	var s SomeType

	funcName := runtime.FuncForPC(reflect.ValueOf((*SomeType).SomeFunc).Pointer()).Name()
	methodName := runtime.FuncForPC(reflect.ValueOf(s.SomeFunc).Pointer()).Name()

	fmt.Printf("funcName: %s\n", funcName)
	fmt.Printf("methodName: %s\n", methodName)
	// Output:
	// 	funcName: main.(*SomeType).SomeFunc
	//  methodName: main.(*SomeType).SomeFunc-fm

}
