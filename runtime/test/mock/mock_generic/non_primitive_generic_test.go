//go:build go1.18
// +build go1.18

package mock_generic

// see bug https://github.com/xhd2015/xgo/issues/211
type GenericSt[T any] struct {
	Data T
}

func (g GenericSt[T]) GetData(param T) T {
	return param
}

type Inner struct {
}
