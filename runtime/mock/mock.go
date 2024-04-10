package mock

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trap"
)

// a marker to indicate the
// original function should be called
var ErrCallOld = errors.New("mock: call old")

type Interceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error

// Mock setup mock on given function `fn`.
// `fn` can be a function or a method,
// if `fn` is a method, only the bound
// instance will be mocked, other instances
// are not affected.
// The returned function can be used to cancel
// the passed interceptor.
func Mock(fn interface{}, interceptor Interceptor) func() {
	recvPtr, fnInfo, funcPC, trappingPC := getFunc(fn)
	return mock(recvPtr, fnInfo, funcPC, trappingPC, interceptor)
}

func MockByName(pkgPath string, funcName string, interceptor Interceptor) func() {
	recv, fn, funcPC, trappingPC := getFuncByName(pkgPath, funcName)
	return mock(recv, fn, funcPC, trappingPC, interceptor)
}

// Can instance be nil?
func MockMethodByName(instance interface{}, method string, interceptor Interceptor) func() {
	recvPtr, fn, funcPC, trappingPC := getMethodByName(instance, method)
	return mock(recvPtr, fn, funcPC, trappingPC, interceptor)
}

func getFunc(fn interface{}) (recvPtr interface{}, fnInfo *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	// if the target function is a method, then a
	// recv ptr must be given
	recvPtr, fnInfo, funcPC, trappingPC = trap.InspectPC(fn)
	if fnInfo == nil {
		pc := reflect.ValueOf(fn).Pointer()
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			panic(fmt.Errorf("failed to setup mock for variable: 0x%x", pc))
		}
		panic(fmt.Errorf("failed to setup mock for: %v", fn.Name()))
	}
	return recvPtr, fnInfo, funcPC, trappingPC
}

func getFuncByName(pkgPath string, funcName string) (recvPtr interface{}, fn *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	fn = functab.GetFuncByPkg(pkgPath, funcName)
	if fn == nil {
		panic(fmt.Errorf("failed to setup mock for: %s.%s", pkgPath, funcName))
	}
	return nil, fn, 0, 0
}

func getMethodByName(instance interface{}, method string) (recvPtr interface{}, fn *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	// extract instance's reflect.Type
	// use that type to query for reflect mapping in functab:
	//    reflectTypeMapping map[reflect.Type]map[string]*funcInfo
	t := reflect.TypeOf(instance)
	funcMapping := functab.GetTypeMethods(t)
	if funcMapping == nil {
		panic(fmt.Errorf("failed to setup mock for type %T", instance))
	}
	fn = funcMapping[method]
	if fn == nil {
		panic(fmt.Errorf("failed to setup mock for: %T.%s", instance, method))
	}

	addr := reflect.New(t)
	addr.Elem().Set(reflect.ValueOf(instance))

	return addr.Interface(), fn, 0, 0
}

// TODO: ensure them run in last?
// no abort, run mocks
// mocks are special in that they on run in pre stage

// if mockFnInfo is a function, mockRecvPtr is always nil
// if mockFnInfo is a method,
//   - if mockRecvPtr has a value, then only call to that instance will be mocked
//   - if mockRecvPtr is nil, then all call to the function will be mocked
func mock(mockRecvPtr interface{}, mockFnInfo *core.FuncInfo, funcPC uintptr, trappingPC uintptr, interceptor Interceptor) func() {
	return trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.Kind == core.Kind_Func && f.PC == 0 {
				if !f.Generic {
					if !f.Closure || trap.ClosureHasFunc {
						return nil, nil
					}
				}
				// may atch generic
				// or closure without PC
			}
			if f != mockFnInfo {
				// no match
				return nil, nil
			}
			if f.Generic && f.RecvType == "" {
				// generic function(not method) should distinguish different implementations
				curTrappingPC := trap.GetTrappingPC()
				if curTrappingPC != 0 && curTrappingPC != funcPC && curTrappingPC != trappingPC {
					return nil, nil
				}
			}

			if f.RecvType != "" && mockRecvPtr != nil {
				// check recv instance
				recvPtr := args.GetFieldIndex(0).Ptr()

				// check they pointing to the same variable
				re := reflect.ValueOf(recvPtr).Elem().Interface()
				me := reflect.ValueOf(mockRecvPtr).Elem().Interface()
				if re != me {
					// if *recvPtr != *mockRecvPtr {
					return nil, nil
				}
			}

			// TODO: add panic check
			err = interceptor(ctx, f, args, result)
			if err != nil {
				if err == ErrCallOld {
					// continue
					return nil, nil
				}
				return nil, err
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
