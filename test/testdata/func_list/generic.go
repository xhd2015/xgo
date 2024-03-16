//go:build go1.18
// +build go1.18

package main

func generic[T any]() {

}

type List[T any] struct {
}

func (c *List[T]) size() int {
	return 0
}
