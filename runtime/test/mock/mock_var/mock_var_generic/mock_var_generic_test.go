//go:build go1.22
// +build go1.22

package mock_var_generic

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func TestMockVarGenericSingleParam(t *testing.T) {
	var called bool
	mock.Mock(&instance, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		called = true
		return nil
	})
	a := instance
	_ = a
	if !called {
		t.Errorf("expect called to be true, actual: %t", called)
	}
}

func TestMockVarGenericDoubleParams(t *testing.T) {
	var called bool
	mock.Mock(&instance2, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		called = true
		return nil
	})
	a := instance2
	_ = a
	if !called {
		t.Errorf("expect called to be true, actual: %t", called)
	}
}
