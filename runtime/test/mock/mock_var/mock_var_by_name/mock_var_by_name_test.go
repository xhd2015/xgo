package mock_var_by_name

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

var a int = 123

const pkgPath = "github.com/xhd2015/xgo/runtime/test/mock/mock_var/mock_var_by_name"

func TestMockVarByNameTest(t *testing.T) {
	mock.MockByName(pkgPath, "a", func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error {
		results.GetFieldIndex(0).Set(456)
		return nil
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
}

func TestMockVarByNamePtrTest(t *testing.T) {
	mock.MockByName(pkgPath, "*a", func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error {
		x := 456
		results.GetFieldIndex(0).Set(&x)
		return nil
	})
	pb := &a
	b := *pb
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
}
