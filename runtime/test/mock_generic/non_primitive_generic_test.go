package mock_generic

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// see bug https://github.com/xhd2015/xgo/issues/211
type GenericSt[T any] struct {
	Data T
}

func (g GenericSt[T]) GetData(param T) T {
	return param
}

type Inner struct {
}

func TestNonPrimitiveGeneric(t *testing.T) {
	v := GenericSt[Inner]{}

	var mocked bool
	mock.Patch(v.GetData, func(Inner) Inner {
		mocked = true
		return Inner{}
	})
	v.GetData(Inner{})
	if !mocked {
		t.Fatalf("expected mocked, actually not")
	}
}
