package trap

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

// a marker to indicate the
// original function should be called
var errCallOld = errors.New("mock: call old")

func MockCallOld() {
	panic(errCallOld)
}

type Interceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error

func PushMockByInterceptor(fn interface{}, interceptor Interceptor) func() {
	fnv := reflect.ValueOf(fn)
	if fnv.Kind() == reflect.Ptr {
		// variable
		handler := func(name string, res interface{}) {
			fnInfo := &core.FuncInfo{
				Kind: core.Kind_Var,
			}
			var argObj object
			resObject := object{
				{
					name:   name,
					valPtr: res,
				},
			}
			interceptor(nil, fnInfo, argObj, resObject)
		}
		return PushVarMockHandler(fnv.Pointer(), handler)
	} else if fnv.Kind() == reflect.Func {
		// func
	} else {
		panic(fmt.Errorf("fn should be func or pointer to variable, actual: %T", fn))
	}

	recvPtr, _, trappingPC := InspectPC(fn)
	handler := buildMockFromInterceptor(recvPtr, interceptor)
	return PushMockHandler(trappingPC, recvPtr, handler)
}

func PushMockByPatch(fn interface{}, replacer interface{}) func() {
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
			return PushVarMockHandler(fnv.Pointer(), handler)
		}
		return PushVarPtrMockHandler(fnv.Pointer(), handler)
	} else if fnv.Kind() == reflect.Func {
		// func
	} else {
		panic(fmt.Errorf("fn should be func or pointer to variable, actual: %T", fn))
	}

	recvPtr, _, trappingPC := InspectPC(fn)
	handler := buildPatchHandler(recvPtr, fn, replacer)
	return PushMockHandler(trappingPC, recvPtr, handler)
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

// pre and post must have the same signature, as:
//
//	fn(args ..., results...)
func buildRecorderHandler(recvPtr interface{}, fn interface{}, pre interface{}, post interface{}) (func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (interface{}, bool), func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}, data interface{})) {
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
		wantArgs = append(wantArgs, reflect.PtrTo(fnType.In(i)))
	}
	for i := 0; i < fnType.NumOut(); i++ {
		wantArgs = append(wantArgs, reflect.PtrTo(fnType.Out(i)))
	}

	wantType := reflect.FuncOf(wantArgs, nil, false)
	if preType != nil && preType != wantType {
		panic(fmt.Errorf("pre should have type: %v, actual: %T", wantType, pre))
	}
	if postType != nil && postType != wantType {
		panic(fmt.Errorf("post should have type: %v, actual: %T", wantType, post))
	}
	nIn := wantType.NumIn()

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	callHandlerHandler := func(fnV reflect.Value, recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (interface{}, bool) {
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
					callArgs[callIdx] = reflect.ValueOf(actRecvPtr)
					callIdx++
				}
			}
		}
		nArgs := len(args)
		for i := 0; i < nArgs; i++ {
			callArgs[callIdx+i] = reflect.ValueOf(args[i])
		}
		nResults := len(results)
		for i := 0; i < nResults; i++ {
			callArgs[callIdx+nArgs+i] = reflect.ValueOf(results[i])
		}

		// call the function
		if !fnType.IsVariadic() {
			fnV.Call(callArgs)
		} else {
			fnV.CallSlice(callArgs)
		}
		return nil, false
	}
	var preHandler func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (interface{}, bool)
	var postHandler func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}, data interface{})

	if pre != nil {
		preHandler = func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (interface{}, bool) {
			return callHandlerHandler(preV, recvName, actRecvPtr, argNames, args, resultNames, results)
		}
	}
	if post != nil {
		postHandler = func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}, data interface{}) {
			callHandlerHandler(postV, recvName, actRecvPtr, argNames, args, resultNames, results)
		}
	}
	return preHandler, postHandler
}

func buildMockFromInterceptor(recvPtr interface{}, interceptor Interceptor) func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) bool {
	if interceptor == nil {
		panic("interceptor is nil")
	}

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	return func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) bool {
		if recvPtr != nil {
			if actRecvPtr == nil {
				panic("receiver mismatch")
			}
		}

		// TODO: fill info
		fnInfo := &core.FuncInfo{
			Kind: core.Kind_Func,
		}
		var argObj object
		var resObject object
		if actRecvPtr != nil {
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

		// call the function
		var callOld bool
		func() {
			defer func() {
				if e := runtime.XgoPeekPanic(); e != nil {
					if pe, ok := e.(error); ok && pe == errCallOld {
						callOld = true
					}
				}
			}()

			interceptor(nil, fnInfo, argObj, resObject)
		}()
		return !callOld
	}
}

func buildRecorderFromInterceptor(recvPtr interface{}, interceptor PreRecordInterceptor, postInterceptor PostRecordInterceptor) (func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (interface{}, bool), func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}, data interface{})) {
	if interceptor == nil {
		panic("interceptor is nil")
	}

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	pre := func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (interface{}, bool) {
		if recvPtr != nil {
			if actRecvPtr == nil {
				panic("receiver mismatch")
			}
		}

		// TODO: fill info
		fnInfo := &core.FuncInfo{
			Kind: core.Kind_Func,
		}
		var argObj object
		var resObject object
		if actRecvPtr != nil {
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
		data, err := interceptor(nil, fnInfo, argObj, resObject)
		if err != nil {
			panic(err)
		}
		return data, false
	}
	post := func(recvName string, actRecvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}, data interface{}) {
		if recvPtr != nil {
			if actRecvPtr == nil {
				panic("receiver mismatch")
			}
		}

		// TODO: fill info
		fnInfo := &core.FuncInfo{
			Kind: core.Kind_Func,
		}
		var argObj object
		var resObject object
		if actRecvPtr != nil {
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
		err := postInterceptor(nil, fnInfo, argObj, resObject, data)
		if err != nil {
			panic(err)
		}
		return
	}
	return pre, post
}

// a,b must be func type
func checkFuncTypeMatch(a reflect.Type, b reflect.Type, skipAFirst bool) (atype string, btype string, match bool) {
	na := a.NumIn()
	nb := b.NumIn()

	base := 0
	if skipAFirst {
		base++
	}
	if na-base != nb {
		return formatFuncType(a, skipAFirst), formatFuncType(b, false), false
	}

	for i := 0; i < na; i++ {
		ta := a.In(i + base)
		tb := b.In(i)
		if ta != tb {
			return formatFuncType(a, skipAFirst), formatFuncType(b, false), false
		}
	}

	nouta := a.NumOut()
	noutb := b.NumOut()
	if nouta != noutb {
		return formatFuncType(a, skipAFirst), formatFuncType(b, false), false
	}
	for i := 0; i < nouta; i++ {
		ta := a.Out(i)
		tb := b.Out(i)
		if ta != tb {
			return formatFuncType(a, skipAFirst), formatFuncType(b, false), false
		}
	}
	return "", "", true
}

func formatFuncType(f reflect.Type, skipFirst bool) string {
	n := f.NumIn()
	i := 0
	if skipFirst {
		i++
	}
	var strBuilder strings.Builder
	strBuilder.WriteString("func(")
	for ; i < n; i++ {
		t := f.In(i)
		strBuilder.WriteString(t.String())
		if i < n-1 {
			strBuilder.WriteString(",")
		}
	}
	strBuilder.WriteString(")")

	nout := f.NumOut()
	if nout > 0 {
		strBuilder.WriteString(" ")
		if nout > 1 {
			strBuilder.WriteString("(")
		}
		for i := 0; i < nout; i++ {
			t := f.Out(i)
			strBuilder.WriteString(t.String())
			if i < nout-1 {
				strBuilder.WriteString(",")
			}
		}
		if nout > 1 {
			strBuilder.WriteString(")")
		}
	}

	return strBuilder.String()
}
