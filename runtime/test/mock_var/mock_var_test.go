package main

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/mock_var/sub"
)

var a int = 123

// TODO: support xgo:notrap
// xgo:notrap
var b int

func TestMockVarTest(t *testing.T) {
	mock.Mock(&a, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set(456)
		return nil
	})
	b := a
	if b != 456 {
		t.Fatalf("expect b to be %d, actual: %d", 456, b)
	}
}

func TestMockVarInOtherPkg(t *testing.T) {
	mock.Mock(&sub.A, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("mockA")
		return nil
	})
	b := sub.A
	if b != "mockA" {
		t.Fatalf("expect sub.A to be %s, actual: %s", "mockA", b)
	}
}
