//go:build go1.18
// +build go1.18

package atomic_generic

var testAtomicPointer func() int

func testLocalPointer() int {
	var p Pointer[int]

	a := 10
	p.Store(&a)

	return *p.Load()
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
