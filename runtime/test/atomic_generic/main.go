package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	var p atomic.Pointer[int64]

	a := int64(10)
	p.Store(&a)

	fmt.Printf("load: %d\n", (*p.Load()))

	testLocalPointer()
}

func testLocalPointer() {
	var p Pointer[int]

	a := 10
	p.Store(&a)

	fmt.Printf("load: %d\n", (*p.Load()))
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
