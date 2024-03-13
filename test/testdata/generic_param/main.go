// expects go1.17 to build fail
// expects go1.18 to build successfully
package main

import "fmt"

var SomeGenericList List[int64]

func main() {
	// ERR: cannot infer V
	// fmt.Println(reverse([]int{1, 2, 3, 4, 5}))

	fmt.Println(reverse[int, int64]([]int{1, 2, 3, 4, 5}))

	// ERR: cannot infer V
	// fmt.Println(reverse[int]([]int{1, 2, 3, 4, 5}))

	// NOTE: generic call does not allow empty []
	// fmt.Println(reverse[]([]int{1, 2, 3, 4, 5}))
}

// struct with generic
type List[T any] struct {
	next  *List[T]
	value T
}

func (l List[T]) CloneLen() int { return 1 }
func (l *List[T]) Len() int { return 1 }

// ERR: generic type cannot be alias
// type List2[T any] = List[T]

// invalid
// type List3=[T any]List[T]

// if uncommented, given compile error with go1.18: function type must have no type parameters
// type X = func[T any](s []T) []T

// func with generic
//
// T is a type parameter that is used like normal type inside the function
// any is a constraint on type i.e T has to implement "any" interface
func reverse[T any, V int64 | float64](s []T) []T {
	l := len(s)
	r := make([]T, l)

	for i, ele := range s {
		r[l-i-1] = ele
	}
	return r
}
// cannot compile, at least one generic param
// func noParam[]()