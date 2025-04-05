package trap

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/internal/generic"
	xgo_runtime "github.com/xhd2015/xgo/runtime/internal/runtime"
)

const methodSuffix = "-fm"

// Inspect inspects the function and returns the receiver pointer, function PC and trapping PC.
// for ordinary function, `funcPC` is the same with `trappingPC`
// for method,`funcPC` and `trappingPC` are different, `funcPC` is the PC of the method, which binds a special pointer to its receiver, and `trappingPC` is the PC of the general type method
func Inspect(f interface{}) (recvPtr interface{}, funcInfo *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	fn := reflect.ValueOf(f)
	if fn.Kind() == reflect.Ptr {
		// a variable
		funcInfo = functab.InfoVarAddr(fn.Pointer())
		if funcInfo == nil {
			panic(fmt.Errorf("variable %w: %v", ErrNotInstrumented, f))
		}
		return nil, funcInfo, 0, 0
	}
	if fn.Kind() != reflect.Func {
		panic(fmt.Errorf("requires func, given: %s", fn.Kind().String()))
	}
	funcPC = fn.Pointer()

	// funcs, closures and type functions can be found directly by PC
	var closureMightBeGeneric bool
	funcInfo = functab.InfoPC(funcPC)
	if funcInfo != nil {
		if !funcInfo.Closure || !generic.GenericImplIsClosure {
			// found funcPC
			return nil, funcInfo, funcPC, funcPC
		}
		// generic
		closureMightBeGeneric = true
	}

	// for go1.18, go1.19, generic implementation
	// is implemented as closure, so we need to distinguish
	// between true closure and generic function
	// full name example: github.com/xhd2015/xgo/runtime/test/trap/inspect.TestInspectGeneric.func1
	// where TestInspectGeneric is the function that uses the generic
	var origFullName string
	var needRecv bool
	if !closureMightBeGeneric {
		origFullName = xgo_runtime.XgoGetFullPCName(funcPC)
		fullName := origFullName

		// check method
		if strings.HasSuffix(fullName, methodSuffix) {
			needRecv = true
			fullName = fullName[:len(fullName)-len(methodSuffix)]
		}

		funcInfo = functab.GetFuncByFullName(fullName)
		if funcInfo == nil {
			// not a method, might be generic or closure
			// we currently cannot handle this
			panic(fmt.Errorf("func %w: %s", ErrNotInstrumented, fullName))
		}

		if funcInfo.Closure && generic.GenericImplIsClosure {
			// maybe closure generic
			closureMightBeGeneric = true
		} else if !needRecv && !funcInfo.Generic {
			// plain function(not method, not generic)
			return nil, funcInfo, funcPC, funcPC
		}
	}

	// retrieve receive ptr
	callFn := func() {
		stack := getOrAttachStackData()
		stack.inspecting = func(pc uintptr, f *core.FuncInfo, actRecvPtr interface{}, args []interface{}, results []interface{}) {
			trappingPC = runtime.FuncForPC(pc).Entry()
			funcInfo = f
			if needRecv || closureMightBeGeneric {
				recvPtr = actRecvPtr
			}
		}
		defer func() {
			stack.inspecting = nil
		}()
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
			origFullName = xgo_runtime.XgoGetFullPCName(funcPC)
		}
		panic(fmt.Errorf("failed to retrieve instance pointer for method: %s", origFullName))
	}
	return recvPtr, funcInfo, funcPC, trappingPC
}
