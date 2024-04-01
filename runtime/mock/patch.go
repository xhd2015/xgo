package mock

import (
	"context"
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core"
)

func PatchByName(pkgPath string, funcName string, replacer interface{}) func() {
	return MockByName(pkgPath, funcName, buildInterceptorFromPatch(replacer))
}

func PatchMethodByName(instance interface{}, method string, replacer interface{}) func() {
	return MockMethodByName(instance, method, buildInterceptorFromPatch(replacer))
}

func buildInterceptorFromPatch(replacer interface{}) func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
	v := reflect.ValueOf(replacer)
	t := v.Type()
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("requires func, given %T", replacer))
	}
	if v.IsNil() {
		panic("replacer is nil")
	}
	nIn := t.NumIn()
	return func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		// assemble arguments
		callArgs := make([]reflect.Value, nIn)
		src := 0
		dst := 0
		if fn.RecvType != "" {
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
		for i := 0; i < nOut; i++ {
			results.GetFieldIndex(i).Set(res[i].Interface())
		}

		// check error
		if nOut > 0 {
			last := res[nOut-1].Interface()
			if last != nil {
				if err, ok := last.(error); ok {
					return err
				}
			}
		}

		return nil
	}
}
