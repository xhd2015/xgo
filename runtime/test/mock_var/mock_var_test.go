package mock_var

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

var c sub.Mapping = sub.Mapping{
	1: "hello",
}

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

func TestThirdPartyTypeMethodVarWithoutWrapShouldNotWork(t *testing.T) {
	mock.Patch(&c, func() sub.Mapping {
		return sub.Mapping{
			1: "mock",
		}
	})
	txt := c.Get(1)
	if txt != "hello" {
		t.Fatalf("expect c[1] to be %s, actual: %s", "hello", txt)
	}
}

func TestThirdPartyTypeMethodVarWithWrapShouldWork(t *testing.T) {
	mock.Patch(&c, func() sub.Mapping {
		return sub.Mapping{
			1: "mock",
		}
	})
	txt := sub.Mapping(c).Get(1)
	if txt != "mock" {
		t.Fatalf("expect c[1] to be %s, actual: %s", "mock", txt)
	}
}
