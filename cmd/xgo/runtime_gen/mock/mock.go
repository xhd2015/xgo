package mock

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/trap"
)

type Interceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error

// a marker to indicate the
// original function should be called
var errCallOld = errors.New("mock: call old")

func CallOld() {
	panic(errCallOld)
}

// Mock setup mock on given function `fn`.
// `fn` can be a function or a method,
// if `fn` is a method, only the bound
// instance will be mocked, other instances
// are not affected.
// The returned function can be used to cancel
// the passed interceptor.
func Mock(fn interface{}, interceptor Interceptor) func() {
	fnv := reflect.ValueOf(fn)
	if fnv.Kind() == reflect.Ptr {
		// variable
		handler := func(name string, res interface{}) {
			fnInfo := &core.FuncInfo{
				Kind: core.Kind_Var,
			}
			var argObj object
			resObject := object{
				{
					name:   name,
					valPtr: res,
				},
			}
			interceptor(nil, fnInfo, argObj, resObject)
		}
		return trap.PushVarMockHandler(fnv.Pointer(), handler)
	} else if fnv.Kind() == reflect.Func {
		// func
	} else {
		panic(fmt.Errorf("fn should be func or pointer to variable, actual: %T", fn))
	}

	recvPtr, _, trappingPC := trap.InspectPC(fn)
	handler := buildMockHandler(recvPtr, fn, interceptor)
	return trap.PushMockHandler(trappingPC, recvPtr, handler)
}

func buildMockHandler(recvPtr interface{}, fn interface{}, interceptor Interceptor) func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) bool {
	if interceptor == nil {
		panic("interceptor is nil")
	}

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	return func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) bool {
		if recvPtr != nil {
			if actRecvPtr == nil {
				panic("receiver mismatch")
			}
		}

		// call the function
		var callOld bool
		func() {
			defer func() {
				if e := runtime.XgoPeekPanic(); e != nil {
					if pe, ok := e.(error); ok && pe == errCallOld {
						callOld = true
					}
				}
			}()
			// TODO: fill info
			fnInfo := &core.FuncInfo{
				Kind: core.Kind_Func,
			}
			var argObj object
			var resObject object
			if actRecvPtr != nil {
				argObj = append(argObj, field{
					name:   recvName,
					valPtr: actRecvPtr,
				})
			}
			for i, arg := range args {
				argObj = append(argObj, field{
					name:   argNames[i],
					valPtr: arg,
				})
			}
			for i, res := range results {
				resObject = append(resObject, field{
					name:   resultNames[i],
					valPtr: res,
				})
			}
			interceptor(nil, fnInfo, argObj, resObject)
		}()
		if callOld {
			return false
		}
		return true
	}
}
