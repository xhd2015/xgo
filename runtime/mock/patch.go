package mock

import (
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/trap"
)

// Patch replaces `fn` with `replacer` in current goroutine.
// You do not have to manually clean up the replacer, as
// xgo will automatically clear the replacer when
// current gorotuine exits.
// However, if you want to clear the replacer earlier,
// this function returns a clean up function that can be
// used to clear the replacer.
func Patch(fn interface{}, replacer interface{}) func() {
	fnv := reflect.ValueOf(fn)
	if fnv.Kind() == reflect.Ptr {
		// variable
		rv := reflect.ValueOf(replacer)
		isPtr := checkVarType(fnv.Type(), rv.Type(), true)
		handler := func(name string, res interface{}) {
			fnRes := rv.Call([]reflect.Value{})
			reflect.ValueOf(res).Elem().Set(fnRes[0])
		}
		if !isPtr {
			return trap.PushVarMockHandler(fnv.Pointer(), handler)
		}
		return trap.PushVarPtrMockHandler(fnv.Pointer(), handler)
	} else if fnv.Kind() == reflect.Func {
		// func
	} else {
		panic(fmt.Errorf("fn should be func or pointer to variable, actual: %T", fn))
	}

	recvPtr, _, trappingPC := trap.InspectPC(fn)
	handler := buildPatchHandler(recvPtr, fn, replacer)
	return trap.PushMockHandler(trappingPC, recvPtr, handler)
}

// return `true` if hit ptr type
func checkVarType(varType reflect.Type, replacerType reflect.Type, supportPtr bool) bool {
	wantValueType := reflect.FuncOf(nil, []reflect.Type{varType.Elem()}, false)
	if replacerType.Kind() != reflect.Func {
		panic(fmt.Errorf("replacer should have type: %s, actual: %s", wantValueType.String(), replacerType.String()))
	}

	targetTypeStr, replacerTypeStr, match := checkFuncTypeMatch(wantValueType, replacerType, false)
	if match {
		return false
	}
	if supportPtr {
		wantPtrType := reflect.FuncOf(nil, []reflect.Type{varType}, false)
		_, _, matchPtr := checkFuncTypeMatch(wantPtrType, replacerType, false)
		if matchPtr {
			return true
		}
	}
	panic(fmt.Errorf("replacer should have type: %s, actual: %s", targetTypeStr, replacerTypeStr))
}

func buildPatchHandler(recvPtr interface{}, fn interface{}, replacer interface{}) func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) bool {
	v := reflect.ValueOf(replacer)
	t := v.Type()
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("requires func, given %T", replacer))
	}
	if v.IsNil() {
		panic("replacer is nil")
	}
	if t != reflect.TypeOf(fn) {
		panic(fmt.Errorf("replacer should have type: %T, actual: %T", fn, replacer))
	}
	nIn := t.NumIn()

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	return func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) bool {
		// assemble arguments
		callArgs := make([]reflect.Value, nIn)
		callIdx := 0

		// if recvPtr is nil, then we can just treat as normal
		// function
		// otherwise, we need to check if actRecvPtr is same as recvPtr

		if recvPtr != nil {
			if actRecvPtr == nil {
				panic("receiver mismatch")
			}
		}

		if actRecvPtr != nil {
			var isInstance bool
			if recvPtr != nil {
				isInstance = true
			}
			if isInstance {
				// patching an instance method
				// replacer's does not have receiver, so no need to fill in
			} else {
				// set receiver
				if nIn > 0 {
					callArgs[callIdx] = reflect.ValueOf(actRecvPtr).Elem()
					callIdx++
				}
			}
		}
		for i := 0; i < nIn-callIdx; i++ {
			// fail if with the following setup:
			//    reflect: Call using zero Value argument
			// callArgs[dst+i] = reflect.ValueOf(args.GetFieldIndex(src + i).Value())
			callArgs[callIdx+i] = reflect.ValueOf(args[i]).Elem()
		}

		// call the function
		var callOld bool
		var res []reflect.Value
		func() {
			defer func() {
				if e := runtime.XgoPeekPanic(); e != nil {
					if pe, ok := e.(error); ok && pe == errCallOld {
						callOld = true
					}
				}
			}()
			if !t.IsVariadic() {
				res = v.Call(callArgs)
			} else {
				res = v.CallSlice(callArgs)
			}
		}()
		if callOld {
			return false
		}

		// assign result
		nOut := len(res)
		resLen := nOut
		for i := 0; i < resLen; i++ {
			reflect.ValueOf(results[i]).Elem().Set(res[i])
		}
		return true
	}
}
