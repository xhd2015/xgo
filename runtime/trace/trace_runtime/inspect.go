package trace_runtime

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

const methodSuffix = "-fm"

// InspectPC inspects the function and returns the receiver pointer, function PC and trapping PC.
// for ordinary function, `funcPC` is the same with `trappingPC`
// for method,`funcPC` and `trappingPC` are different, `funcPC` is the PC of the method, and `trappingPC` is the PC of the general type method
func InspectPC(f interface{}) (recvPtr interface{}, funcPC uintptr, trappingPC uintptr) {
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func {
		panic(fmt.Errorf("Inspect requires func, given: %s", fn.Kind().String()))
	}
	funcPC = fn.Pointer()

	fullName := Runtime_XgoGetFullPCName(funcPC)
	if !strings.HasSuffix(fullName, methodSuffix) {
		return nil, funcPC, funcPC
	}

	// retrieve recieve ptr
	callFn := func() {
		stack := GetOrAttachStack()
		stack.inspecting = func(pc uintptr, recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) {
			trappingPC = runtime.FuncForPC(pc).Entry()
			recvPtr = actRecvPtr
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
	return recvPtr, funcPC, trappingPC
}
