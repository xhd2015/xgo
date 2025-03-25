package mock

import (
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/trace/trace_runtime"
)

// Patch replaces `fn` with `replacer` in current goroutine.
// You do not have to manually clean up the replacer, as
// xgo will automatically clear the replacer when
// current gorotuine exits.
// However, if you want to clear the replacer earlier,
// this function returns a clean up function that can be
// used to clear the replacer.
func Patch(fn interface{}, replacer interface{}) func() {
	recvPtr, _, trappingPC := trace_runtime.InspectPC(fn)
	handler := buildMockHandler(recvPtr, replacer)
	return trace_runtime.PushMockHandler(trappingPC, recvPtr, handler)
}

func buildMockHandler(recvPtr interface{}, replacer interface{}) func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) {
	v := reflect.ValueOf(replacer)
	t := v.Type()
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("requires func, given %T", replacer))
	}
	if v.IsNil() {
		panic("replacer is nil")
	}
	nIn := t.NumIn()

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	return func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) {
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
		var res []reflect.Value
		if !t.IsVariadic() {
			res = v.Call(callArgs)
		} else {
			res = v.CallSlice(callArgs)
		}

		// assign result
		nOut := len(res)
		resLen := nOut
		for i := 0; i < resLen; i++ {
			reflect.ValueOf(results[i]).Elem().Set(res[i])
		}
	}
}
