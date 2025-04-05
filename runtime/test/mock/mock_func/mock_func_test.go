package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/core/info"
	"github.com/xhd2015/xgo/runtime/mock"
)

// go run ./cmd/xgo run --project-dir ./runtime ./test/mock
func MockFuncTest(t *testing.T) {

	mock.Mock(hello, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("mock")
		return nil
	})
	x := hello("world")
	if x != "mock" {
		t.Fatalf("expect hello(%q) to be %q, actual: %q", "world", "mock", x)
	}

	// should fail on skip
	var havePanicked bool
	func() {
		defer func() {
			if e := recover(); e != nil {
				havePanicked = true
			}
		}()
		mock.Mock(hello_skipped, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
			return nil
		})
	}()
	if !havePanicked {
		t.Fatalf("expect mock on hello_skipped should panic, actual not panic")
	}
}

func hello(a string) string {
	return fmt.Sprintf("hello %s\n", a)
}

func hello_skipped(a string) string {
	panic("hello_skipped")

	return fmt.Sprintf("hello %s skipped\n", a)
}

func neverErr() error {
	return nil
}

func TestMockFuncErr(t *testing.T) {
	mockErr := errors.New("mock err")
	mock.Mock(neverErr, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return mockErr
	})
	err := neverErr()
	if err != mockErr {
		t.Fatalf("expect mocked neverErr() to be %v, actual: %v", mockErr, err)
	}
}

func TestNestedMock(t *testing.T) {
	// before mock
	beforeMock := A()
	beforeMockExpect := "A B"
	if beforeMock != beforeMockExpect {
		t.Fatalf("expect before mock: %q, actual: %q", beforeMockExpect, beforeMock)
	}
	mock.Mock(B, func(ctx context.Context, fn *info.Func, args, results core.Object) error {
		results.GetFieldIndex(0).Set("b")
		return nil
	})
	mock.Mock(A, func(ctx context.Context, fn *info.Func, args, results core.Object) error {
		results.GetFieldIndex(0).Set("a " + B())
		return nil
	})
	afterMock := A()
	afterMockExpect := "a b"
	if afterMock != afterMockExpect {
		t.Fatalf("expect after mock: %q, actual: %q", afterMockExpect, afterMock)
	}
}

func A() string {
	return "A " + B()
}
func B() string {
	return "B"
}
