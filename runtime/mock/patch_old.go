package mock

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trap"
)

// Patch replaces `fn` with `replacer` in current goroutine.
// You do not have to manually clean up the replacer, as
// xgo will automatically clear the replacer when
// current gorotuine exits.
// However, if you want to clear the replacer earlier,
// this function returns a clean up function that can be
// used to clear the replacer.
// Deprecated: use Patch instead.
func PatchV1(fn interface{}, replacer interface{}) func() {
	if fn == nil {
		panic("fn cannot be nil")
	}
	if replacer == nil {
		panic("replacer cannot be nil")
	}
	fnType := reflect.TypeOf(fn)
	fnKind := fnType.Kind()
	if fnKind == reflect.Func {
		if fnType != reflect.TypeOf(replacer) {
			panic(fmt.Errorf("replacer should have type: %T, actual: %T", fn, replacer))
		}
	} else if fnKind == reflect.Ptr {
		replacerType := reflect.TypeOf(replacer)
		wantType := reflect.FuncOf(nil, []reflect.Type{fnType.Elem()}, false)
		var targetTypeStr string
		var replacerTypeStr string
		var match bool
		if reflect.TypeOf(replacer).Kind() != reflect.Func {
			targetTypeStr = wantType.String()
			replacerTypeStr = replacerType.String()
		} else {
			targetTypeStr, replacerTypeStr, match = checkFuncTypeMatch(wantType, replacerType, false)
		}
		if !match {
			panic(fmt.Errorf("replacer should have type: %s, actual: %s", targetTypeStr, replacerTypeStr))
		}
	} else {
		panic(fmt.Errorf("fn should be func or pointer to variable, actual: %T", fn))
	}

	recvPtr, fnInfo, funcPC, trappingPC := getFunc(fn)
	return mock(recvPtr, fnInfo, funcPC, trappingPC, buildInterceptorFromPatch(recvPtr, replacer))
}

func PatchByName(pkgPath string, funcName string, replacer interface{}) func() {
	if replacer == nil {
		panic("replacer cannot be nil")
	}
	t := reflect.TypeOf(replacer)
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("replacer should be func, actual: %T", replacer))
	}

	// check type
	recvPtr, funcInfo, funcPC, trappingPC := getFuncByName(pkgPath, funcName)
	if funcInfo.Kind == core.Kind_Func {
		if funcInfo.Func != nil {
			calledType, replacerType, match := checkFuncTypeMatch(reflect.TypeOf(funcInfo.Func), t, recvPtr != nil)
			if !match {
				panic(fmt.Errorf("replacer should have type: %s, actual: %s", calledType, replacerType))
			}
		}
	} else if funcInfo.Kind == core.Kind_Var || funcInfo.Kind == core.Kind_VarPtr || funcInfo.Kind == core.Kind_Const {
		varPtrType := reflect.TypeOf(funcInfo.Var)
		var wantValueType reflect.Type
		if funcInfo.Kind == core.Kind_Var {
			wantValueType = reflect.FuncOf(nil, []reflect.Type{varPtrType.Elem()}, false)
		} else {
			// const: type is not pointer
			wantValueType = reflect.FuncOf(nil, []reflect.Type{varPtrType}, false)
		}

		var targetTypeStr string
		var replacerTypeStr string
		var match bool

		var matchPtr bool
		replacerType := reflect.TypeOf(replacer)
		if replacerType.Kind() != reflect.Func {
			targetTypeStr = wantValueType.String()
			replacerTypeStr = replacerType.String()
		} else {
			targetTypeStr, replacerTypeStr, match = checkFuncTypeMatch(wantValueType, replacerType, false)
			if !match && funcInfo.Kind != core.Kind_VarPtr {
				_, replacerTypeStr, match = checkFuncTypeMatch(reflect.FuncOf(nil, []reflect.Type{varPtrType}, false), replacerType, false)
				matchPtr = true
			}
		}
		if !match {
			panic(fmt.Errorf("replacer should have type: %s, actual: %s", targetTypeStr, replacerTypeStr))
		}
		if matchPtr {
			funcInfo = functab.Info(pkgPath, "*"+funcName)
			if funcInfo == nil {
				panic(fmt.Errorf("failed to patch: %s *%s", pkgPath, funcName))
			}
		}
	} else {
		panic(fmt.Errorf("unrecognized func type: %s", funcInfo.Kind.String()))
	}

	return mock(recvPtr, funcInfo, funcPC, trappingPC, buildInterceptorFromPatch(recvPtr, replacer))
}

func PatchMethodByName(instance interface{}, method string, replacer interface{}) func() {
	if replacer == nil {
		panic("replacer cannot be nil")
	}
	t := reflect.TypeOf(replacer)
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("replacer should be func, actual: %T", replacer))
	}

	// check type
	recvPtr, funcInfo, funcPC, trappingPC := getMethodByName(instance, method)
	if funcInfo.Func != nil {
		calledType, replacerType, match := checkFuncTypeMatch(reflect.TypeOf(funcInfo.Func), t, recvPtr != nil)
		if !match {
			panic(fmt.Errorf("replacer should have type: %s, actual: %s", calledType, replacerType))
		}
	}
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

	// replacer is usually a closure,
	// we can bypass it
	trap.Ignore(replacer)

	// first arg ctx: true => [recv,args[1:]...]
	// first arg ctx: false => [recv, args[0:]...]
	return func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		// assemble arguments
		callArgs := make([]reflect.Value, nIn)
		src := 0
		dst := 0

		if fn.RecvType != "" {
			var isInstance bool
			if recvPtr != nil {
				if fn.Generic && trap.GenericImplIsClosure && args.NumField() == nIn {
					// not an instance
				} else {
					isInstance = true
				}
			}
			if isInstance {
				// patching an instance method
				src++
				// replacer's does not have receiver
			} else {
				// set receiver
				if nIn > 0 {
					callArgs[dst] = reflect.ValueOf(args.GetFieldIndex(0).Ptr()).Elem()
					dst++
					src++
				}
			}
		}
		if fn.FirstArgCtx {
			callArgs[dst] = reflect.ValueOf(ctx)
			dst++
		}
		for i := 0; i < nIn-dst; i++ {
			// fail if with the following setup:
			//    reflect: Call using zero Value argument
			// callArgs[dst+i] = reflect.ValueOf(args.GetFieldIndex(src + i).Value())
			callArgs[dst+i] = reflect.ValueOf(args.GetFieldIndex(src + i).Ptr()).Elem()
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
