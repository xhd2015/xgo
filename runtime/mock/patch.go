package mock

import (
	"context"
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core"
)

func PatchByName(pkgPath string, funcName string, replacer interface{}) func() {
	recvPtr, funcInfo, funcPC, trappingPC := getFuncByName(pkgPath, funcName)
	return mock(recvPtr, funcInfo, funcPC, trappingPC, buildInterceptorFromPatch(recvPtr, replacer))
}

func PatchMethodByName(instance interface{}, method string, replacer interface{}) func() {
	recvPtr, funcInfo, funcPC, trappingPC := getMethodByName(instance, method)
	return mock(recvPtr, funcInfo, funcPC, trappingPC, buildInterceptorFromPatch(recvPtr, replacer))
}

func buildInterceptorFromPatch(recvPtr interface{}, replacer interface{}) func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
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
	return func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		// assemble arguments
		callArgs := make([]reflect.Value, nIn)
		src := 0
		dst := 0
		if fn.RecvType != "" && recvPtr != nil {
			// patching an instance method
			src++
		}
		if fn.FirstArgCtx {
			callArgs[0] = reflect.ValueOf(ctx)
			dst++
		}
		for i := 0; i < nIn-dst; i++ {
			callArgs[dst+i] = reflect.ValueOf(args.GetFieldIndex(src + i).Value())
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
		if fn.LastResultErr {
			resLen--
		}
		for i := 0; i < resLen; i++ {
			results.GetFieldIndex(i).Set(res[i].Interface())
		}

		if fn.LastResultErr {
			results.(core.ObjectWithErr).GetErr().Set(res[nOut-1].Interface())
		}

		return nil
	}
}
