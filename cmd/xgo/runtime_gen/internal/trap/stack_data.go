package trap

import (
	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/stack"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

type dataKeyType struct{}

var dataKey = dataKeyType{}

var globalInterceptorHolder interceptorHolders

type StackData struct {
	// when handling trapping, prohibit
	// another trapping from the handler
	handlingTrapping bool

	hasStartedTracing bool

	onFinish        func(stack stack_model.IStack)
	stackOutputFile string

	inspecting func(pc uintptr, funcInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{})

	interceptors interceptorHolders
}

type interceptorHolders struct {
	// pc->mock
	mock       map[uintptr][]*mockHolder
	varMock    map[uintptr][]*varMockHolder
	varPtrMock map[uintptr][]*varMockHolder

	// pc->recorder
	recorder       map[uintptr][]*recorderHolder
	varRecorder    map[uintptr][]*varRecordHolder
	varPtrRecorder map[uintptr][]*varRecordHolder

	// general purpose interceptors
	interceptors []*recorderHolder
}

func getStackData() *StackData {
	return getStackDataOf(stack.Get())
}

func getOrAttachStackData() *StackData {
	return getOrAttachStackDataOf(stack.GetOrAttach())
}

func getStackDataOf(stk *stack.Stack) *StackData {
	if stk == nil {
		return nil
	}
	data := stk.GetData(dataKey)
	d, _ := data.(*StackData)
	return d
}

func getOrAttachStackDataOf(stk *stack.Stack) *StackData {
	data := stk.GetData(dataKey)
	d, ok := data.(*StackData)
	if ok {
		return d
	}
	d = &StackData{}
	stk.SetData(dataKey, d)
	return d
}

type recorderHolder struct {
	wantRecvPtr interface{}
	pre         func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) (interface{}, bool)
	post        func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}, data interface{})
}

type varRecordHolder struct {
	pre  func(fnInfo *core.FuncInfo, res interface{}) (interface{}, bool)
	post func(fnInfo *core.FuncInfo, res interface{}, data interface{})
}

func (c *StackData) getLastMock(pc uintptr) (recvPtr interface{}, mock func(fnInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{}) bool) {
	mockList := c.interceptors.mock[pc]
	if len(mockList) == 0 {
		mockList = globalInterceptorHolder.mock[pc]
		if len(mockList) == 0 {
			return nil, nil
		}
	}
	m := mockList[len(mockList)-1]
	return m.wantRecvPtr, m.mock
}

func (c *StackData) getLastVarMock(varAddr uintptr) (mock func(fnInfo *core.FuncInfo, res interface{})) {
	mockList := c.interceptors.varMock[varAddr]
	if len(mockList) == 0 {
		mockList = globalInterceptorHolder.varMock[varAddr]
		if len(mockList) == 0 {
			return nil
		}
	}
	m := mockList[len(mockList)-1]
	return m.mock
}

func (c *StackData) getLastVarPtrMock(varAddr uintptr) (mock func(fnInfo *core.FuncInfo, res interface{})) {
	mockList := c.interceptors.varPtrMock[varAddr]
	if len(mockList) == 0 {
		mockList = globalInterceptorHolder.varPtrMock[varAddr]
		if len(mockList) == 0 {
			return nil
		}
	}
	m := mockList[len(mockList)-1]
	return m.mock
}

func (c *StackData) getRecordHandlers(pc uintptr) []*recorderHolder {
	globalRecorders := globalInterceptorHolder.recorder[pc]
	localRecorders := c.interceptors.recorder[pc]
	if len(globalRecorders) > 0 && len(localRecorders) > 0 {
		list := make([]*recorderHolder, len(localRecorders)+len(globalRecorders))
		copy(list, localRecorders)
		copy(list[len(localRecorders):], globalRecorders)
		return list
	}
	if len(localRecorders) > 0 {
		return localRecorders
	}
	return globalRecorders
}

func (c *StackData) getVarRecordHandlers(varAddr uintptr) []*varRecordHolder {
	globalVarRecorders := globalInterceptorHolder.varRecorder[varAddr]
	localVarRecorders := c.interceptors.varRecorder[varAddr]
	if len(globalVarRecorders) > 0 && len(localVarRecorders) > 0 {
		list := make([]*varRecordHolder, len(localVarRecorders)+len(globalVarRecorders))
		copy(list, localVarRecorders)
		copy(list[len(localVarRecorders):], globalVarRecorders)
		return list
	}
	if len(localVarRecorders) > 0 {
		return localVarRecorders
	}
	return globalVarRecorders
}

func (c *StackData) getVarPtrRecordHandlers(varAddr uintptr) []*varRecordHolder {
	globalVarPtrRecorders := globalInterceptorHolder.varPtrRecorder[varAddr]
	localVarPtrRecorders := c.interceptors.varPtrRecorder[varAddr]
	if len(globalVarPtrRecorders) > 0 && len(localVarPtrRecorders) > 0 {
		list := make([]*varRecordHolder, len(localVarPtrRecorders)+len(globalVarPtrRecorders))
		copy(list, localVarPtrRecorders)
		copy(list[len(localVarPtrRecorders):], globalVarPtrRecorders)
		return list
	}
	if len(localVarPtrRecorders) > 0 {
		return localVarPtrRecorders
	}
	return globalVarPtrRecorders
}

func (c *StackData) getGeneralInterceptors() []*recorderHolder {
	globalInterceptors := globalInterceptorHolder.interceptors
	localInterceptors := c.interceptors.interceptors
	if len(globalInterceptors) > 0 && len(localInterceptors) > 0 {
		list := make([]*recorderHolder, len(localInterceptors)+len(globalInterceptors))
		copy(list, localInterceptors)
		copy(list[len(localInterceptors):], globalInterceptors)
		return list
	}
	if len(localInterceptors) > 0 {
		return localInterceptors
	}
	return globalInterceptors
}
