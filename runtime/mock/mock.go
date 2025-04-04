package mock

import (
	"context"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/trap"
)

type Interceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error

// Mock setup mock on given function `fn`.
// `fn` can be a function or a method,
// if `fn` is a method, only the bound
// instance will be mocked, other instances
// are not affected.
// The returned function can be used to cancel
// the passed interceptor.
func Mock(fn interface{}, interceptor Interceptor) func() {
	return trap.PushMockInterceptor(fn, trap.Interceptor(interceptor))
}

func MockByName(pkgPath string, funcName string, interceptor Interceptor) func() {
	return trap.PushMockByName(pkgPath, funcName, trap.Interceptor(interceptor))
}

func MockMethodByName(instance interface{}, method string, interceptor Interceptor) func() {
	return trap.PushMockMethodByName(instance, method, trap.Interceptor(interceptor))
}
