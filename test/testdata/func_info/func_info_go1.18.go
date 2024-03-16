//go:build go1.18
// +build go1.18

package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/functab"
)

func runGeneric() {
	fmt.Printf("generic func info\n")

	var list List[int]
	list.Size()

	Hello(list)

	funcInfo := functab.Info("main", "(*List).Size")
	if funcInfo == nil {
		panic(fmt.Errorf("func (*List).Size not found"))
	}

	fmt.Printf("(*List).Size identityName: %s\n", funcInfo.IdentityName)
	fmt.Printf("(*List).Size args: %v\n", funcInfo.ArgNames)

	funcInfoHello := functab.Info("main", "Hello")
	if funcInfoHello == nil {
		panic(fmt.Errorf("func Hello not found"))
	}
	fmt.Printf("Hello identityName: %s\n", funcInfoHello.IdentityName)
	fmt.Printf("Hello args: %v\n", funcInfoHello.ArgNames)
}

type List[T any] struct {
}

func (c *List[T]) Size() {

}

// funcName1: Hello[go.shape.struct {}]
// funcName2: Hello[main.List[int]]
func Hello[T any](v T) interface{} {
	return v
}
