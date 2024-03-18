package mock

import (
	"context"
	"errors"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

var ErrCallOld = errors.New("mock: call old")

type Interceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error

var interceptors []Interceptor

func GetInterceptors() []Interceptor {
	return interceptors
}

// TODO: ensure them run in last?
// no abort, run mocks
// mocks are special in that they on run in pre stage
func AddFuncInterceptor(fn interface{}, interceptor Interceptor) {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if !f.IsFunc(fn) {
				// continue
				return nil, nil
			}
			// TODO: add panic check
			err = interceptor(ctx, f, args, result)
			if err == ErrCallOld {
				// continue
				return nil, nil
			}
			// when match func, default to use mock
			return nil, trap.ErrAbort
		},
	})
}

func CallOld() {
	// TODO: implement recover
	panic(ErrCallOld)
}

// mock context
// MockContext
// MockPoint
// MockController
// MockRecorder
// type MockContext interface {
// 	Log()
// }

// a mock context seems not needed, because all can be done by calling
// into mock's function
