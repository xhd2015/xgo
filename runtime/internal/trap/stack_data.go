package trap

import (
	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/stack"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

type dataKeyType struct{}

var dataKey = dataKeyType{}

type StackData struct {
	// when handling trapping, prohibit
	// another trapping from the handler
	handlingTrapping bool

	hasStartedTracing bool

	onFinish        func(stack stack_model.IStack)
	stackOutputFile string

	inspecting func(pc uintptr, funcInfo *core.FuncInfo, recvPtr interface{}, args []interface{}, results []interface{})

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
	mockList := c.mock[pc]
	if len(mockList) == 0 {
		return nil, nil
	}
	m := mockList[len(mockList)-1]
	return m.wantRecvPtr, m.mock
}

func (c *StackData) getLastVarMock(varAddr uintptr) (mock func(fnInfo *core.FuncInfo, res interface{})) {
	mockList := c.varMock[varAddr]
	if len(mockList) == 0 {
		return nil
	}
	m := mockList[len(mockList)-1]
	return m.mock
}

func (c *StackData) getLastVarPtrMock(varAddr uintptr) (mock func(fnInfo *core.FuncInfo, res interface{})) {
	mockList := c.varPtrMock[varAddr]
	if len(mockList) == 0 {
		return nil
	}
	m := mockList[len(mockList)-1]
	return m.mock
}

func (c *StackData) getRecordHandlers(pc uintptr) []*recorderHolder {
	return c.recorder[pc]
}

func (c *StackData) getVarRecordHandlers(varAddr uintptr) []*varRecordHolder {
	return c.varRecorder[varAddr]
}

func (c *StackData) getVarPtrRecordHandlers(varAddr uintptr) []*varRecordHolder {
	return c.varPtrRecorder[varAddr]
}
