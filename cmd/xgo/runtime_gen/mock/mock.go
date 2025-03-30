package mock

import (
	"github.com/xhd2015/xgo/runtime/trap"
)

func CallOld() {
	trap.MockCallOld()
}

// Mock setup mock on given function `fn`.
// `fn` can be a function or a method,
// if `fn` is a method, only the bound
// instance will be mocked, other instances
// are not affected.
// The returned function can be used to cancel
// the passed interceptor.
func Mock(fn interface{}, interceptor trap.Interceptor) func() {
	return trap.PushMockByInterceptor(fn, interceptor)
}
