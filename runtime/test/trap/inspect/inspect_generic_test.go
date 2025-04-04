//go:build go1.18
// +build go1.18

package inspect

func GenericFunc[T any]() {
}

type GenericType[T any] struct {
}

func (g *GenericType[T]) GenericMethod() {
}
