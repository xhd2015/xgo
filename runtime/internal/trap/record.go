package trap

import (
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

func PushRecorder(fn interface{}, pre interface{}, post interface{}) func() {
	return pushRecorder(fn, pre, post)
}

func pushRecorder(fn interface{}, pre interface{}, post interface{}) func() {
	fnv := reflect.ValueOf(fn)
	if fnv.Kind() == reflect.Ptr {
		varPtr := fnv.Pointer()
		funcInfo := functab.InfoVarAddr(varPtr)
		if funcInfo == nil {
			panic(fmt.Errorf("variable %w: %v", ErrNotInstrumented, varPtr))
		}
		if pre == nil && post == nil {
			panic(fmt.Errorf("pre and post should not be both nil"))
		}

		if pre != nil && post != nil && reflect.TypeOf(pre) != reflect.TypeOf(post) {
			panic(fmt.Errorf("pre-recorder and post-recorder should have the same type, actual: pre has %T, post has %T", pre, post))
		}

		// variable
		var preHandler func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool)
		var postHandler func(fnInfo *core.FuncInfo, res interface{}, data interface{})

		var preIsPtr bool
		var postIsPtr bool

		if pre != nil {
			preV := reflect.ValueOf(pre)
			var preRecordArgTypes []ptrType
			preRecordArgTypes, preIsPtr = checkVarRecorderType(fnv.Type(), preV.Type(), true)
			preHandler = func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool) {
				arg := preRecordArgTypes[0].get(reflect.ValueOf(res))
				preV.Call([]reflect.Value{arg})
				return nil, false
			}
		}
		if post != nil {
			postV := reflect.ValueOf(post)
			var postRecordArgTypes []ptrType
			postRecordArgTypes, postIsPtr = checkVarRecorderType(fnv.Type(), postV.Type(), true)
			postHandler = func(fnInfo *core.FuncInfo, res interface{}, data interface{}) {
				arg := postRecordArgTypes[0].get(reflect.ValueOf(res))
				postV.Call([]reflect.Value{arg})
			}
		}
		if pre != nil && post != nil && preIsPtr == postIsPtr {
			panic(fmt.Errorf("pre-recorder and post-recorder should have the same type, actual: pre has %T, post has %T", pre, post))
		}

		if !preIsPtr {
			return PushVarRecordHandler(varPtr, preHandler, postHandler)
		}
		return PushVarPtrRecordHandler(varPtr, preHandler, postHandler)
	} else if fnv.Kind() == reflect.Func {
		// func
	} else {
		panic(fmt.Errorf("fn should be func or pointer to variable, actual: %T", fn))
	}

	recvPtr, _, _, trappingPC := Inspect(fn)
	preHandler, postHandler := buildRecorderHandler(recvPtr, fn, pre, post)
	return PushRecordHandler(trappingPC, recvPtr, preHandler, postHandler)
}

func pushRecorderInterceptor(fn interface{}, preInterceptor PreInterceptor, postInterceptor PostInterceptor) func() {
	fnv := reflect.ValueOf(fn)
	if fnv.Kind() == reflect.Ptr {
		varPtr := fnv.Pointer()
		funcInfo := functab.InfoVarAddr(varPtr)
		if funcInfo == nil {
			panic(fmt.Errorf("variable %w: %v", ErrNotInstrumented, varPtr))
		}
		// variable
		preHandler := func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool) {
			var argObj object
			resObject := object{
				{
					name:   fnInfo.Name,
					valPtr: res,
				},
			}
			data, _ := preInterceptor(nil, funcInfo, argObj, resObject)
			return data, false
		}
		postHandler := func(fnInfo *core.FuncInfo, res interface{}, data interface{}) {
			var argObj object
			resObject := object{
				{
					name:   fnInfo.Name,
					valPtr: res,
				},
			}
			postInterceptor(nil, funcInfo, argObj, resObject, data)
		}
		return PushVarRecordHandler(varPtr, preHandler, postHandler)
	} else if fnv.Kind() == reflect.Func {
		// func
	} else {
		panic(fmt.Errorf("fn should be func or pointer to variable, actual: %T", fn))
	}

	recvPtr, _, _, trappingPC := Inspect(fn)
	pre, post := buildRecorderFromInterceptor(recvPtr, preInterceptor, postInterceptor)
	return PushRecordHandler(trappingPC, recvPtr, pre, post)
}

func PushRecordHandler(pc uintptr, recvPtr interface{}, pre func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool), post func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}, data interface{})) func() {
	stack := getOrAttachStackData()
	if stack.recorder == nil {
		stack.recorder = map[uintptr][]*recorderHolder{}
	}
	h := &recorderHolder{wantRecvPtr: recvPtr, pre: pre, post: post}
	stack.recorder[pc] = append(stack.recorder[pc], h)
	return func() {
		list := stack.recorder[pc]
		n := len(list)
		if list[n-1] == h {
			stack.recorder[pc] = list[:n-1]
			return
		}
		// remove at some index
		for i, m := range list {
			if m == h {
				stack.recorder[pc] = append(list[:i], list[i+1:]...)
				return
			}
		}
		panic(fmt.Errorf("pop recorder not found, check if the recorder is already popped earlier"))
	}
}

func PushVarRecordHandler(varAddr uintptr, pre func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool), post func(fnInfo *core.FuncInfo, res interface{}, data interface{})) func() {
	stack := getOrAttachStackData()
	if stack.varRecorder == nil {
		stack.varRecorder = map[uintptr][]*varRecordHolder{}
	}
	h := &varRecordHolder{pre: pre, post: post}
	stack.varRecorder[varAddr] = append(stack.varRecorder[varAddr], h)
	return func() {
		list := stack.varRecorder[varAddr]
		n := len(list)
		if list[n-1] == h {
			stack.varRecorder[varAddr] = list[:n-1]
			return
		}
		// remove at some index
		for i, m := range list {
			if m == h {
				stack.varRecorder[varAddr] = append(list[:i], list[i+1:]...)
				return
			}
		}
		panic(fmt.Errorf("pop recorder not found, check if the recorder is already popped earlier"))
	}
}

func PushVarPtrRecordHandler(varAddr uintptr, pre func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool), post func(fnInfo *core.FuncInfo, res interface{}, data interface{})) func() {
	stack := getOrAttachStackData()
	if stack.varPtrRecorder == nil {
		stack.varPtrRecorder = map[uintptr][]*varRecordHolder{}
	}
	h := &varRecordHolder{pre: pre, post: post}
	stack.varPtrRecorder[varAddr] = append(stack.varPtrRecorder[varAddr], h)
	return func() {
		list := stack.varPtrRecorder[varAddr]
		n := len(list)
		if list[n-1] == h {
			stack.varPtrRecorder[varAddr] = list[:n-1]
			return
		}
		// remove at some index
		for i, m := range list {
			if m == h {
				stack.varPtrRecorder[varAddr] = append(list[:i], list[i+1:]...)
				return
			}
		}
		panic(fmt.Errorf("pop recorder not found, check if the recorder is already popped earlier"))
	}
}

func checkVarRecorderType(varPtrType reflect.Type, recorderType reflect.Type, supportPtr bool) ([]ptrType, bool) {
	varType := varPtrType.Elem()
	printWantType := reflect.FuncOf([]reflect.Type{varType}, nil, false)
	if recorderType.Kind() != reflect.Func {
		panic(fmt.Errorf("recorder should have type: %v, actual: %s", printWantType, recorderType.String()))
	}
	recordArgTypes, ok := resolveArgTypes(recorderType, []reflect.Type{varType})
	if ok {
		return recordArgTypes, false
	}
	if supportPtr {
		recordArgTypes, ok := resolveArgTypes(recorderType, []reflect.Type{varPtrType})
		if ok {
			return recordArgTypes, true
		}
	}
	panic(fmt.Errorf("recorder should have type: %v, actual: %T", printWantType, recorderType))
}
