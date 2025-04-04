package trap

import (
	"context"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/trap"
)

type Interceptor struct {
	Pre  func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object) (data interface{}, err error)
	Post func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object, data interface{}) error
}

// AddInterceptor add a general interceptor
func AddInterceptor(interceptor *Interceptor) func() {
	return trap.PushInterceptor(interceptor.Pre, interceptor.Post)
}

// AddFuncInterceptor add func interceptor
func AddFuncInterceptor(f interface{}, interceptor *Interceptor) func() {
	return trap.PushRecorderInterceptor(f, interceptor.Pre, interceptor.Post)
}
