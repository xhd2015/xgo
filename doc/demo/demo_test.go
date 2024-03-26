package demo

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func MyFunc() string {
	return "my func"
}
func TestFuncMock(t *testing.T) {
	mock.Mock(MyFunc, func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error {
		results.GetFieldIndex(0).Set("mock func")
		return nil
	})
	text := MyFunc()
	if text != "mock func" {
		t.Fatalf("expect MyFunc() to be 'mock func', actual: %s", text)
	}
}
