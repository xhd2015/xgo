package trap

import (
	"context"
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core"
)

func buildMockHandler(recvPtr interface{}, funcInfo *core.FuncInfo, replacer interface{}) func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) bool {
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
	return func(fnInfo *core.FuncInfo, actRecvPtr interface{}, args []interface{}, results []interface{}) bool {
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
		return true
	}
}

// pre and post must have the same signature, as:
//
//	fn(args ..., results...)
func buildRecorderHandler(recvPtr interface{}, fn interface{}, pre interface{}, post interface{}) (func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool), func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}, data interface{})) {
	if pre == nil && post == nil {
		panic("pre and post are nil")
	}
	var preV reflect.Value
	var preType reflect.Type
	if pre != nil {
		preV = reflect.ValueOf(pre)
		preType = preV.Type()
		if preType.Kind() != reflect.Func {
			panic(fmt.Errorf("requires pre to be func, given %T", pre))
		}
	}

	var postV reflect.Value
	var postType reflect.Type
	if post != nil {
		postV = reflect.ValueOf(post)
		postType = postV.Type()
		if postType.Kind() != reflect.Func {
			panic(fmt.Errorf("requires post to be func, given %T", post))
		}
	}

	fnType := reflect.TypeOf(fn)
	wantArgs := make([]reflect.Type, 0, fnType.NumIn()+fnType.NumOut())
	for i := 0; i < fnType.NumIn(); i++ {
		wantArgs = append(wantArgs, fnType.In(i))
	}
	for i := 0; i < fnType.NumOut(); i++ {
		wantArgs = append(wantArgs, fnType.Out(i))
	}

	var preArgTypes []ptrType
	var postArgTypes []ptrType
	printWantType := reflect.FuncOf(wantArgs, nil, false)
	if preType != nil {
		var ok bool
		preArgTypes, ok = resolveArgTypes(preType, wantArgs)
		if !ok {
			panic(fmt.Errorf("pre should have type: %v, actual: %T", printWantType, pre))
		}
	}
	if postType != nil {
		var ok bool
		postArgTypes, ok = resolveArgTypes(postType, wantArgs)
		if !ok {
			panic(fmt.Errorf("post should have type: %v, actual: %T", printWantType, post))
		}
	}

	nIn := printWantType.NumIn()

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	callHandlerHandler := func(fnInfo *core.FuncInfo, fnV reflect.Value, fnType reflect.Type, argTypes []ptrType, actRecvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool) {
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
					callArgs[callIdx] = argTypes[callIdx].get(reflect.ValueOf(actRecvPtr))
					callIdx++
				}
			}
		}
		nArgs := len(args)
		for i := 0; i < nArgs; i++ {
			callArgs[callIdx] = argTypes[callIdx].get(reflect.ValueOf(args[i]))
			callIdx++
		}
		nResults := len(results)
		for i := 0; i < nResults; i++ {
			callArgs[callIdx] = argTypes[callIdx].get(reflect.ValueOf(results[i]))
			callIdx++
		}

		// call the function
		if !fnType.IsVariadic() {
			fnV.Call(callArgs)
		} else {
			fnV.CallSlice(callArgs)
		}
		return nil, false
	}
	var preHandler func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool)
	var postHandler func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}, data interface{})

	if pre != nil {
		preHandler = func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool) {
			return callHandlerHandler(fnInfo, preV, preType, preArgTypes, recvPtr, args, results)
		}
	}
	if post != nil {
		postHandler = func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}, data interface{}) {
			callHandlerHandler(fnInfo, postV, postType, postArgTypes, recvPtr, args, results)
		}
	}
	return preHandler, postHandler
}

func buildMockFromInterceptor(recvPtr interface{}, interceptor Interceptor) func(funcInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) bool {
	if interceptor == nil {
		panic("interceptor is nil")
	}

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	return func(funcInfo *core.FuncInfo, actRecvPtr interface{}, args []interface{}, results []interface{}) bool {
		if recvPtr != nil {
			if actRecvPtr == nil {
				panic("receiver mismatch")
			}
		}

		recvName := funcInfo.RecvName
		argNames := funcInfo.ArgNames
		resultNames := funcInfo.ResNames

		var argObj object
		var resObject object
		if recvPtr == nil && actRecvPtr != nil {
			// see https://github.com/xhd2015/xgo/issues/316
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

		var ctx context.Context = context.TODO()
		if funcInfo.FirstArgCtx {
			// see https://github.com/xhd2015/xgo/issues/316
			ptr := args[0].(*context.Context)
			ctx = *ptr
		}

		err := interceptor(ctx, funcInfo, argObj, resObject)
		if err != nil {
			if funcInfo.LastResultErr {
				lastErr := results[len(results)-1].(*error)
				*lastErr = err
			} else {
				panic(err)
			}
		}
		return true
	}
}

func buildRecorderFromInterceptor(recvPtr interface{}, preInterceptor PreInterceptor, postInterceptor PostInterceptor) (func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool), func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}, data interface{})) {
	if preInterceptor == nil && postInterceptor == nil {
		panic("pre and post interceptor are nil")
	}
	var pre func(funcInfo *core.FuncInfo, actRecvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool)
	var post func(funcInfo *core.FuncInfo, actRecvPtr interface{}, args []interface{}, results []interface{}, data interface{})

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	if preInterceptor != nil {
		pre = func(funcInfo *core.FuncInfo, actRecvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool) {
			if recvPtr != nil {
				if actRecvPtr == nil {
					panic("receiver mismatch")
				}
			}

			var argObj object
			var resObject object
			if recvPtr == nil && actRecvPtr != nil {
				argObj = append(argObj, field{
					name:   funcInfo.RecvName,
					valPtr: actRecvPtr,
				})
			}
			for i, arg := range args {
				argObj = append(argObj, field{
					name:   funcInfo.ArgNames[i],
					valPtr: arg,
				})
			}
			for i, res := range results {
				resObject = append(resObject, field{
					name:   funcInfo.ResNames[i],
					valPtr: res,
				})
			}
			var ctx context.Context = context.TODO()
			if funcInfo.FirstArgCtx {
				ptr := args[0].(*context.Context)
				ctx = *ptr
			}
			data, err := preInterceptor(ctx, funcInfo, argObj, resObject)
			if err != nil {
				if err == ErrMocked {
					return nil, true
				}
				if funcInfo.LastResultErr {
					lastErr := results[len(results)-1].(*error)
					*lastErr = err
				} else {
					panic(err)
				}
			}
			return data, false
		}
	}
	if postInterceptor != nil {
		post = func(funcInfo *core.FuncInfo, actRecvPtr interface{}, args []interface{}, results []interface{}, data interface{}) {
			if recvPtr != nil {
				if actRecvPtr == nil {
					panic("receiver mismatch")
				}
			}

			var argObj object
			var resObject object
			if recvPtr == nil && actRecvPtr != nil {
				argObj = append(argObj, field{
					name:   funcInfo.RecvName,
					valPtr: actRecvPtr,
				})
			}
			for i, arg := range args {
				argObj = append(argObj, field{
					name:   funcInfo.ArgNames[i],
					valPtr: arg,
				})
			}
			for i, res := range results {
				resObject = append(resObject, field{
					name:   funcInfo.ResNames[i],
					valPtr: res,
				})
			}
			var ctx context.Context = context.TODO()
			if funcInfo.FirstArgCtx {
				ptr := args[0].(*context.Context)
				ctx = *ptr
			}
			err := postInterceptor(ctx, funcInfo, argObj, resObject, data)
			if err != nil {
				if funcInfo.LastResultErr {
					lastErr := results[len(results)-1].(*error)
					*lastErr = err
				} else {
					panic(err)
				}
			}
		}
	}
	return pre, post
}
