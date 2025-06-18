//go:build go1.22
// +build go1.22

package mock_var_generic

// see Issue https://github.com/xhd2015/xgo/issues/357

type Wrapper[T any] struct {
}

type Wrapper2[T any, V any] struct {
}

type Concrete struct {
}
type Concrete2 struct {
}

var instance = Wrapper[Concrete]{}

var instance2 = Wrapper2[Concrete, Concrete2]{}
