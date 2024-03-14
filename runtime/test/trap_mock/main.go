package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/core/trace"

	"github.com/xhd2015/xgo/runtime/test/trap/mock"
)

func init() {
	trace.Use()
	mock.Use()
}

func main() {
	A()
	B()
	C()
}

func A() {
	fmt.Printf("A\n")
}

func B() {
	fmt.Printf("B\n")
	C()
}
func C() {
	fmt.Printf("C\n")
}
