package trap

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

const methodSuffix = "-fm"

// Inspect make a call to f to capture its receiver pointer if it
// is bound method
// It can be used to get the unwrapped innermost function of a method
// wrapper.
// if f is a bound method, then guaranteed that recvPtr cannot be nil
func Inspect(f interface{}) (recvPtr interface{}, funcInfo *core.FuncInfo) {
	recvPtr, funcInfo, _, _ = InspectPC(f)
	return
}

type inspectingFunc func(f *core.FuncInfo, recv interface{}, pc uintptr)

func InspectPC(f interface{}) (recvPtr interface{}, funcInfo *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	fn := reflect.ValueOf(f)
	// try as a variable
	if fn.Kind() == reflect.Ptr {
		// a variable
		funcInfo = functab.InfoVar(f)
		return nil, funcInfo, 0, 0
	}
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
	var needRecv bool
	if !maybeClosureGeneric {
		origFullName = __xgo_link_get_pc_name(funcPC)
		fullName := origFullName

		if strings.HasSuffix(fullName, methodSuffix) {
			needRecv = true
			fullName = fullName[:len(fullName)-len(methodSuffix)]
		}

		funcInfo = functab.GetFuncByFullName(fullName)
		if funcInfo == nil {
			return nil, nil, funcPC, 0
		}
		if funcInfo.Closure && GenericImplIsClosure {
			// maybe closure generic
			maybeClosureGeneric = true
		} else if !needRecv && !funcInfo.Generic {
			// plain function(not method, not generic)
			return nil, funcInfo, funcPC, 0
		}
	}

	key := uintptr(__xgo_link_getcurg())
	ensureTrapInstall()
	inspectingMap.Store(key, inspectingFunc(func(f *core.FuncInfo, recv interface{}, pc uintptr) {
		trappingPC = pc
		funcInfo = f
		if needRecv || maybeClosureGeneric {
			// closure cannot have receiver pointer
			recvPtr = recv
		}
	}))
	defer inspectingMap.Delete(key)
	callFn := func() {
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
	}
	callFn()
	if needRecv && recvPtr == nil {
		if origFullName == "" {
			origFullName = __xgo_link_get_pc_name(funcPC)
		}
		panic(fmt.Errorf("failed to retrieve instance pointer for method: %s", origFullName))
	}
	return recvPtr, funcInfo, funcPC, trappingPC
}
