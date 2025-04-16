package trap

import (
	"context"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/trap"
)

// ErrMocked indicates the target function is mocked
// and not called, instead result is set by the interceptor
// it is usually used to setup general mock interceptors
// example usage:
//
//		trap.AddInterceptor(&trap.Interceptor{
//			Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
//				if f.IdentityName == "Read" {
//	                result.GetFieldIndex(0).SetInt(100)
//					return nil, trap.ErrMocked
//				}
//				return nil, nil
//			},
//		})
//
// NOTE: this error should only be returned
// by interceptor in pre-phase. If it is
// returned in post-phase, it will cause
// xgo to panic.
var ErrMocked = trap.ErrMocked

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

// MarkIntercept marks given function or variable to be intercepted,
// it serves as a compiler hint to xgo. it does not have
// any runtime behavior.
//
// NOTE: you don't have to explicitly call this function if you
// called these functions in other places within the main module:
// - `mock.Patch(fn,...)`, `mock.Mock(fn,...)`
// - `trace.Record(fn,...)`, `trace.RecordCall(fn,...)`
// - `trap.AddFuncInterceptor(fn,...)`
// - `functab.InfoFunc(fn)`, `functab.InfoVar(addr)`
//
// Example:
//
//	trap.MarkIntercept((*bytes.Buffer).String)
func MarkIntercept(f interface{}) {}

// MarkInterceptAll marks all functions to be intercepted
// this effectively enables the --trap-all flag unless
// explicitly disabled by --trap-all=false flag
func MarkInterceptAll() {}
