package trap

import (
	"context"
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
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
		// variable
		if preInterceptor == nil && postInterceptor == nil {
			panic(fmt.Errorf("preInterceptor and postInterceptor should not be both nil"))
		}
		varPtr := fnv.Pointer()
		funcInfo := functab.InfoVarAddr(varPtr)
		if funcInfo == nil {
			panic(fmt.Errorf("variable %w: %v", ErrNotInstrumented, varPtr))
		}
		var preHandler func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool)
		var postHandler func(fnInfo *core.FuncInfo, res interface{}, data interface{})

		if preInterceptor != nil {
			preHandler = func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool) {
				var argObj object
				resObject := object{
					{
						name:   fnInfo.Name,
						valPtr: res,
					},
				}
				data, err := preInterceptor(context.Background(), funcInfo, argObj, resObject)
				if err != nil {
					if err == ErrMocked {
						return nil, true
					}
					panic(err)
				}

				return data, false
			}
		}
		if postInterceptor != nil {
			postHandler = func(fnInfo *core.FuncInfo, res interface{}, data interface{}) {
				var argObj object
				resObject := object{
					{
						name:   fnInfo.Name,
						valPtr: res,
					},
				}
				err := postInterceptor(context.Background(), funcInfo, argObj, resObject, data)
				if err != nil {
					panic(err)
				}
			}
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
	holder := &globalInterceptorHolder
	if runtime.XgoInitFinished() {
		stack := getOrAttachStackData()
		holder = &stack.interceptors
	}
	if holder.recorder == nil {
		holder.recorder = map[uintptr][]*recorderHolder{}
	}
	h := &recorderHolder{wantRecvPtr: recvPtr, pre: pre, post: post}
	holder.recorder[pc] = append(holder.recorder[pc], h)
	return func() {
		if holder == &globalInterceptorHolder && runtime.XgoInitFinished() {
			panic("global recorder cannot be cancelled after init finished")
		}
		list := holder.recorder[pc]
		n := len(list)
		if list[n-1] == h {
			holder.recorder[pc] = list[:n-1]
			return
		}
		// remove at some index
		for i, m := range list {
			if m == h {
				holder.recorder[pc] = append(list[:i], list[i+1:]...)
				return
			}
		}
		panic(fmt.Errorf("pop recorder not found, check if the recorder is already popped earlier"))
	}
}

func PushVarRecordHandler(varAddr uintptr, pre func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool), post func(fnInfo *core.FuncInfo, res interface{}, data interface{})) func() {
	holder := &globalInterceptorHolder
	if runtime.XgoInitFinished() {
		stack := getOrAttachStackData()
		holder = &stack.interceptors
	}
	if holder.varRecorder == nil {
		holder.varRecorder = map[uintptr][]*varRecordHolder{}
	}
	h := &varRecordHolder{pre: pre, post: post}
	holder.varRecorder[varAddr] = append(holder.varRecorder[varAddr], h)
	return func() {
		if holder == &globalInterceptorHolder && runtime.XgoInitFinished() {
			panic("global recorder cannot be cancelled after init finished")
		}
		list := holder.varRecorder[varAddr]
		n := len(list)
		if list[n-1] == h {
			holder.varRecorder[varAddr] = list[:n-1]
			return
		}
		// remove at some index
		for i, m := range list {
			if m == h {
				holder.varRecorder[varAddr] = append(list[:i], list[i+1:]...)
				return
			}
		}
		panic(fmt.Errorf("pop recorder not found, check if the recorder is already popped earlier"))
	}
}

func PushVarPtrRecordHandler(varAddr uintptr, pre func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool), post func(fnInfo *core.FuncInfo, res interface{}, data interface{})) func() {
	holder := &globalInterceptorHolder
	if runtime.XgoInitFinished() {
		stack := getOrAttachStackData()
		holder = &stack.interceptors
	}
	if holder.varPtrRecorder == nil {
		holder.varPtrRecorder = map[uintptr][]*varRecordHolder{}
	}
	h := &varRecordHolder{pre: pre, post: post}
	holder.varPtrRecorder[varAddr] = append(holder.varPtrRecorder[varAddr], h)
	return func() {
		if holder == &globalInterceptorHolder && runtime.XgoInitFinished() {
			panic("global recorder cannot be cancelled after init finished")
		}
		list := holder.varPtrRecorder[varAddr]
		n := len(list)
		if list[n-1] == h {
			holder.varPtrRecorder[varAddr] = list[:n-1]
			return
		}
		// remove at some index
		for i, m := range list {
			if m == h {
				holder.varPtrRecorder[varAddr] = append(list[:i], list[i+1:]...)
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
		panic(fmt.Errorf("recorder should have type: `%v`, actual: `%s`", printWantType, recorderType.String()))
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
	panic(fmt.Errorf("recorder should have type: %v, actual: %v", printWantType, recorderType))
}
