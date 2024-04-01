package trap

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

const methodSuffix = "-fm"

// Inspect make a call to f to capture its receiver pointer if is
// is bound method
// It can be used to get the unwrapped innermost function of a method
// wrapper.
// if f is a bound method, then guaranteed that recvPtr cannot be nil
func Inspect(f interface{}) (recvPtr interface{}, funcInfo *core.FuncInfo) {
	recvPtr, funcInfo, _, _ = InspectPC(f)
	return
}

func InspectPC(f interface{}) (recvPtr interface{}, funcInfo *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func {
		panic(fmt.Errorf("Inspect requires func, given: %s", fn.Kind().String()))
	}
	funcPC = fn.Pointer()

	// funcs, closures and type functions can be found directly by PC
	var maybeClosureGeneric bool
	funcInfo = functab.InfoPC(funcPC)
	if funcInfo != nil {
		if !funcInfo.Closure || !GenericImplIsClosure {
			return nil, funcInfo, funcPC, 0
		}
		maybeClosureGeneric = true
	}

	// for go1.18, go1.19, generic implementation
	// is implemented as closure, so we need to distinguish
	// between true closure and generic function
	var origFullName string
	if !maybeClosureGeneric {
		origFullName = __xgo_link_get_pc_name(funcPC)
		fullName := origFullName

		var isMethod bool
		if strings.HasSuffix(fullName, methodSuffix) {
			isMethod = true
			fullName = fullName[:len(fullName)-len(methodSuffix)]
		}

		funcInfo = functab.GetFuncByFullName(fullName)
		if funcInfo == nil {
			return nil, nil, funcPC, 0
		}
		if funcInfo.Closure && GenericImplIsClosure {
			// maybe closure generic
			maybeClosureGeneric = true
		} else if !isMethod && !funcInfo.Generic {
			// plain function(not method, not generic)
			return nil, funcInfo, funcPC, 0
		}
	}

	WithInterceptor(&Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			trappingPC = GetTrappingPC()
			funcInfo = f
			if !maybeClosureGeneric {
				// closure cannot have receiver pointer
				recvPtr = args.GetFieldIndex(0).Ptr()
			}
			return nil, ErrAbort
		},
	}, func() {
		fnType := fn.Type()
		nargs := fnType.NumIn()
		args := make([]reflect.Value, nargs)
		for i := 0; i < nargs; i++ {
			args[i] = reflect.New(fnType.In(i)).Elem()
		}
		if !fnType.IsVariadic() {
			fn.Call(args)
		} else {
			fn.CallSlice(args)
		}
	})
	if !maybeClosureGeneric && recvPtr == nil {
		if origFullName == "" {
			origFullName = __xgo_link_get_pc_name(funcPC)
		}
		panic(fmt.Errorf("failed to retrieve instance pointer for method: %s", origFullName))
	}
	return recvPtr, funcInfo, funcPC, trappingPC
}
