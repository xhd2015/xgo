package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/trace"
)

func init() {
	trace.Enable()
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
