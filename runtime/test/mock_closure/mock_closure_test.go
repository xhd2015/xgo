package mock_closuer

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

var gc = func() {
	panic("gc should be mocked")
}

// go run ./cmd/xgo test --project-dir runtime -run TestMockClosure -v ./test/mock_closure
// go run ./script/run-test/ --include go1.22.1 --xgo-runtime-test-only -run TestMockClosure -v ./test/mock_closure
func TestMockClosure(t *testing.T) {
	var lc = func() {
		panic("lc should be mocked")
	}
	mock.Mock(gc, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})
	gc()

	mock.Mock(lc, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		return nil
	})
	lc()
}
