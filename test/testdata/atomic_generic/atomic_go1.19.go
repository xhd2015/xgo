//go:build go1.19
// +build go1.19

package main

import (
	"fmt"
	"sync/atomic"
)

func init() {
	testAtomicPointer = testAtomicPointerGo119
}
func testAtomicPointerGo119() {
	var p atomic.Pointer[int64]

	a := int64(10)
	p.Store(&a)

	fmt.Printf("atomic sload: %d\n", (*p.Load()))
}
