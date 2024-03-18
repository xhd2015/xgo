//go:build go1.18
// +build go1.18

package main

import (
	"fmt"
)

var testAtomicPointer func()

func main() {
	if testAtomicPointer != nil {
		testAtomicPointer()
	}
	testLocalPointer()
}

func testLocalPointer() {
	var p Pointer[int]

	a := 10
	p.Store(&a)

	fmt.Printf("local load: %d\n", (*p.Load()))
}

type Pointer[T any] struct {
	a *int
}

func (c *Pointer[T]) Store(a *int) {
	c.a = a
}

func (c *Pointer[T]) Load() *int {
	return c.a
}
