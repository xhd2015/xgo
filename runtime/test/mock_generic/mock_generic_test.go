//go:build go1.18
// +build go1.18

package mock_generic

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func ToString[T any](v T) string {
	return fmt.Sprint(v)
}

// go run ./cmd/xgo test --project-dir runtime -run TestMockGenericFunc -v ./test/mock_generic
func TestMockGenericFunc(t *testing.T) {
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

type Formatter[T any] struct {
	prefix T
}

func (c *Formatter[T]) Format(v T) string {
	return fmt.Sprintf("%v: %v", c.prefix, v)
}

// go run ./cmd/xgo test --project-dir runtime -run TestMockGenericMethod -v ./test/mock_generic
func TestMockGenericMethod(t *testing.T) {
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
