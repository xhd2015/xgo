//go:build go1.19
// +build go1.19

package atomic_generic

import (
	"sync/atomic"
)

func init() {
	testAtomicPointer = testAtomicPointerGo119
}

func testAtomicPointerGo119() int {
	var p atomic.Pointer[int64]

	a := int64(10)
	p.Store(&a)

	return int(*p.Load())
}
