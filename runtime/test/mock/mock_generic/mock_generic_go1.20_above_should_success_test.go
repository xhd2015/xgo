//go:build go1.20
// +build go1.20

package mock_generic

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func TestMockGenericFuncGo120AboveShouldSuccess(t *testing.T) {
	mock.Mock(ToString[int], func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("hello world")
		return nil
	})
	expect := "hello world"
	output := ToString[int](0)
	if expect != output {
		t.Fatalf("expect ToString[int](0) to be %s, actual: %s", expect, output)
	}

	// per t param
	expectStr := "my string"
	outputStr := ToString[string](expectStr)
	if expectStr != outputStr {
		t.Fatalf("expect ToString[string](%q) not affected, actual: %q", expectStr, outputStr)
	}
}

func TestMockGenericMethodGo120AboveShouldSuccess(t *testing.T) {
	formatter := &Formatter[int]{prefix: 1}
	mock.Mock(formatter.Format, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("hello world")
		return nil
	})
	expect := "hello world"
	output := formatter.Format(0)
	if expect != output {
		t.Fatalf("expect ToString[int](0) to be %s, actual: %s", expect, output)
	}
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
