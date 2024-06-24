// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code
//
// usage:
//  go run -tags dev ./cmd/xgo test --project-dir runtime/test/debug
//  go run -tags dev ./cmd/xgo test --debug-compile --project-dir runtime/test/debug

package debug

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

func TestNonPrimitiveGenericAllInstance(t *testing.T) {
	var mocked bool
	mock.Patch(GenericSt[Inner].GetData, func(GenericSt[Inner], Inner) Inner {
		mocked = true
		return Inner{}
	})
	v := GenericSt[Inner]{}
	v.GetData(Inner{})
	if !mocked {
		t.Fatalf("expected mocked, actually not")
	}
}
