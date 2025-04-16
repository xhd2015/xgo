package trap

import (
	"encoding/json"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	xgo_runtime "github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/internal/stack"
)

func trapVar(infoPtr unsafe.Pointer, varAddr interface{}, res interface{}) {
	funcInfo := (*core.FuncInfo)(infoPtr)

	stk := stack.Get()
	if stk == stack.NilGStack {
		return
	}
	stkData := getStackDataOf(stk)
	ptr := reflect.ValueOf(varAddr).Pointer()

	recorders := stkData.getVarRecordHandlers(ptr)
	mock := stkData.getLastVarMock(ptr)

	depth := xgo_runtime.GetG().IncTrappingDepth()
	defer xgo_runtime.GetG().DecTrappingDepth()

	var tracing bool
	var interceptors []*recorderHolder
	if depth <= 1 {
		if stkData != nil && stkData.hasStartedTracing {
			tracing = true
		}
		interceptors = stkData.getGeneralInterceptors()
	}

	if mock == nil && len(recorders) == 0 && len(interceptors) == 0 && !tracing {
		return
	}
	begin := xgo_runtime.XgoRealTimeNow()
	doTrapVar(funcInfo, stk, begin, tracing, res, recorders, interceptors, mock, res)
}

func trapVarPtr(infoPtr unsafe.Pointer, varAddr interface{}, res interface{}) {
	funcInfo := (*core.FuncInfo)(infoPtr)

	stk := stack.Get()
	if stk == stack.NilGStack {
		return
	}
	stkData := getStackDataOf(stk)
	ptr := reflect.ValueOf(varAddr).Pointer()

	recorders := stkData.getVarPtrRecordHandlers(ptr)
	// var_ptr fallback to var is buggy, can affect program correctness
	// because variable will be overridden
	// and the original variable value will be lost.
	// so we disable it.
	const DISABLE_PTR_FALLBACK = true

	mockRes := res
	mock := stkData.getLastVarPtrMock(ptr)
	if mock == nil && !DISABLE_PTR_FALLBACK {
		mock = stkData.getLastVarMock(ptr)
		if mock != nil {
			// input  res: **T

			// create a temporary variable to hold the value
			rvRes := reflect.ValueOf(res)
			rvVarValue := rvRes.Elem().Elem()

			// new(T)
			tmpRes := reflect.New(rvVarValue.Type())
			tmpRes.Elem().Set(rvVarValue)

			// change res to point to the new variable
			rvRes.Elem().Set(tmpRes)

			// output mockRes: *T
			mockRes = tmpRes.Interface()
		}
	}

	depth := xgo_runtime.GetG().IncTrappingDepth()
	defer xgo_runtime.GetG().DecTrappingDepth()

	var tracing bool
	var interceptors []*recorderHolder
	if depth <= 1 {
		if stkData != nil && stkData.hasStartedTracing {
			tracing = true
		}
		interceptors = stkData.getGeneralInterceptors()
	}
	if mock == nil && len(recorders) == 0 && len(interceptors) == 0 && !tracing {
		return
	}
	begin := xgo_runtime.XgoRealTimeNow()
	doTrapVar(funcInfo, stk, begin, tracing, res, recorders, interceptors, mock, mockRes)
}

func doTrapVar(funcInfo *core.FuncInfo, stk *stack.Stack, begin time.Time, tracing bool, res interface{}, recorders []*varRecordHolder, interceptors []*recorderHolder, mock func(fnInfo *core.FuncInfo, res interface{}), mockRes interface{}) {
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

	if !tracing {
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
