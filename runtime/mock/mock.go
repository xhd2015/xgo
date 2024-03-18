package mock

import (
	"context"
	"errors"
	"sync"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

var ErrCallOld = errors.New("mock: call old")
var errContinue = ErrCallOld

type Interceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error

var interceptors []Interceptor

func GetInterceptors() []Interceptor {
	return interceptors
}

// func AddInterceptor(interceptor Interceptor) {
// 	ensureTrap()
// 	interceptors = append(interceptors, interceptor)
// }

var once sync.Once

func ensureTrap() {
	// TODO: ensure them run in last?
	// no abort, run mocks
	// mocks are special in that they on run in pre stage
	once.Do(func() {
		trap.AddInterceptor(&trap.Interceptor{
			Pre: runMockInterceptors,
		})
	})
}

func runMockInterceptors(ctx context.Context, f *core.FuncInfo, arg, result core.Object) (data interface{}, err error) {
	for _, interceptor := range interceptors {
		// TODO: add panic check
		err := interceptor(ctx, f, arg, result)
		if err == errContinue {
			// call old: let the control flow
			// goes to next interceptor until old function
			continue
		}
		// interrupt mocks and following trap interceptors
		return nil, trap.ErrAbort
	}
	return
}

func AddFuncInterceptor(fn interface{}, interceptor Interceptor) {
	ensureTrap()
	interceptors = append(interceptors, func(ctx context.Context, funcInfo *core.FuncInfo, args, results core.Object) error {
		if !funcInfo.IsFunc(fn) {
			return errContinue
		}
		// when match func, default to use mock
		return interceptor(ctx, funcInfo, args, results)
	})
	// AddInterceptor(func(ctx context.Context, funcInfo *core.FuncInfo, args core.Object, results core.Object) error {
	// 	if !funcInfo.IsFunc(fn) {
	// 		return interceptor(ctx, funcInfo, args, results)
	// 	}
	// 	return nil
	// })
}

func CallOld() {
	// TODO: implement recover
	panic(ErrCallOld)
}

// mock context
// MockContext
// MockTree
// MockEntry
// MockPoint
// MockData
// MockEnv
// MockController
// MockControl
// MockRegistry
// MockManager
// MockCtx
// Mock
// MockFure
// MockA
// Action
// MockLogger
// MockRecorder
// MockDynamic
// MockShim
// MockMod
// MockHandler
// MockRecorder
// Recorder
// type MockContext interface {
// 	Log()
// }

// a mock context seems not needed, because all can be done by calling
// into mock's function
