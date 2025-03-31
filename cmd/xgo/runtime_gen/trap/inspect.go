package trap

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/runtime/functab"
	xgo_runtime "github.com/xhd2015/xgo/runtime/internal/runtime"
)

const methodSuffix = "-fm"

// ErrNotInstrumented is the error returned when a function or variable
// when this error happened, user should check if `--trap` was correctly
// passed to the xgo compiler
// for variable, only variables within main module are available,
// and the variable should be declared with type specified:
//
//	var SomeVar int = 10 // good
//	var SomeVar = 10 // bad: xgo cannot generate trap function for untyped variables
//
// and when used as method receiver, should wrap the variable in type conversion:
//
//	MyInt(SomeVar).String() // good
//	SomeVar.String() // bad: xgo cannot determine if it is a pointer receiver
var ErrNotInstrumented = errors.New("not instrumented by xgo")

// InspectPC inspects the function and returns the receiver pointer, function PC and trapping PC.
// for ordinary function, `funcPC` is the same with `trappingPC`
// for method,`funcPC` and `trappingPC` are different, `funcPC` is the PC of the method, which binds a special pointer to its receiver, and `trappingPC` is the PC of the general type method
func InspectPC(f interface{}) (recvPtr interface{}, funcPC uintptr, trappingPC uintptr) {
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func {
		panic(fmt.Errorf("requires func, given: %s", fn.Kind().String()))
	}
	funcPC = fn.Pointer()

	fullName := xgo_runtime.XgoGetFullPCName(funcPC)
	if !strings.HasSuffix(fullName, methodSuffix) {
		return nil, funcPC, funcPC
	}
	funcInfo := functab.GetFuncByFullName(fullName[:len(fullName)-len(methodSuffix)])
	if funcInfo == nil {
		panic(fmt.Errorf("func %w: %s", ErrNotInstrumented, fullName))
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
