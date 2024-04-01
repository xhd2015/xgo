package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/trap"
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
	trap.Skip()

	return fmt.Sprintf("hello %s skipped\n", a)
}
