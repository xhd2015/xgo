package trap

import (
	"encoding/json"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core/info"
	xgo_runtime "github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/internal/stack"
)

func trapVar(infoPtr unsafe.Pointer, varAddr interface{}, res interface{}) {
	funcInfo := (*info.Func)(infoPtr)

	begin := xgo_runtime.XgoRealTimeNow()
	stk := stack.Get()
	if stk == nil {
		return
	}
	stkData := getStackDataOf(stk)
	if stkData == nil {
		return
	}

	ptr := reflect.ValueOf(varAddr).Pointer()
	recorders := stkData.getVarRecordHandlers(ptr)

	mock := stkData.getLastVarMock(ptr)
	doTrapVar(funcInfo, stk, stkData, begin, res, recorders, mock, res)
}

func trapVarPtr(infoPtr unsafe.Pointer, varAddr interface{}, res interface{}) {
	funcInfo := (*info.Func)(infoPtr)

	begin := xgo_runtime.XgoRealTimeNow()
	stk := stack.Get()
	if stk == nil {
		return
	}

	stkData := getStackDataOf(stk)
	if stkData == nil {
		return
	}

	ptr := reflect.ValueOf(varAddr).Pointer()
	recorders := stkData.getVarPtrRecordHandlers(ptr)

	mockRes := res
	mock := stkData.getLastVarPtrMock(ptr)
	if mock == nil {
		mock = stkData.getLastVarMock(ptr)
		if mock != nil {
			// res: **T
			// mockRes: *T
			mockRes = reflect.ValueOf(res).Elem().Interface()
		}
	}

	doTrapVar(funcInfo, stk, stkData, begin, res, recorders, mock, mockRes)
}

func doTrapVar(funcInfo *info.Func, stk *stack.Stack, stkData *StackData, begin time.Time, res interface{}, recorders []*varRecordHolder, mock func(fnInfo *info.Func, res interface{}), mockRes interface{}) {
	var postRecorders []func()
	for _, recorder := range recorders {
		var data interface{}
		if recorder.pre != nil {
			data, _ = recorder.pre(funcInfo, res)
		}
		if recorder.post != nil {
			postRecorders = append(postRecorders, func() {
				recorder.post(funcInfo, res, data)
			})
		}
	}

	var postInterceptors []func()
	interceptors := stkData.interceptors
	for _, interceptor := range interceptors {
		var data interface{}
		if interceptor.pre != nil {
			data, _ = interceptor.pre(funcInfo, nil, nil, []interface{}{res})
		}
		if interceptor.post != nil {
			postInterceptors = append(postInterceptors, func() {
				interceptor.post(funcInfo, nil, nil, []interface{}{res}, data)
			})
		}
	}

	if mock != nil {
		mock(funcInfo, mockRes)
	}

	for _, recorder := range postRecorders {
		recorder()
	}
	for _, interceptor := range postInterceptors {
		interceptor()
	}

	if !stkData.hasStartedTracing {
		return
	}
	_, file, line, _ := runtime.Caller(SKIP + 2)
	cur := stk.NewEntry(begin, funcInfo.Name)
	cur.File = file
	cur.Line = line
	cur.FuncInfo = funcInfo
	cur.Results = json.RawMessage(xgo_runtime.MarshalNoError(res))
	stk.Top = stk.Push(cur)
	cur.EndNs = xgo_runtime.XgoRealTimeNow().UnixNano() - stk.Begin.UnixNano()
	cur.Finished = true
}
