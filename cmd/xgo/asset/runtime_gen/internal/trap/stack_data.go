package trap

import (
	"sync"
	"sync/atomic"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/internal/stack"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

type dataKeyType struct{}

var dataKey = dataKeyType{}

var globalInterceptorHolder interceptorHolders

// globalInterceptorHolderUsed is 1 once anything is registered into
// globalInterceptorHolder. That only happens during init, but goroutines
// started by init functions may already run instrumented code, so it is
// accessed atomically.
var globalInterceptorHolderUsed int32

// markGlobalInterceptorHolderUsed is called after enabling the trap for
// what was registered: a caller that sees the holder used must also find
// the trap enabled. Variables need nothing enabled: XgoTrapVar and
// XgoTrapVarPtr are called on every access and ignore the skip.
func markGlobalInterceptorHolderUsed() {
	atomic.StoreInt32(&globalInterceptorHolderUsed, 1)
}

func isGlobalInterceptorHolderUsed() bool {
	return atomic.LoadInt32(&globalInterceptorHolderUsed) != 0
}

// trapEnabled is 1 once enableTrap was called. Registrations can race, so
// it is accessed atomically; trapSkipMu orders the first enableTrap with
// init, which sets the skip only if no registration happened before it.
var trapEnabled int32
var trapSkipMu sync.Mutex

// enableTrap makes instrumented functions call trap again. Until then they
// skip it, which is only valid while nothing has been registered and no
// trace or inspection was started, i.e. trap would return immediately.
// Once enabled it stays enabled.
func enableTrap() {
	if atomic.LoadInt32(&trapEnabled) != 0 {
		return
	}
	trapSkipMu.Lock()
	defer trapSkipMu.Unlock()
	// clear the skip before publishing the flag: a caller that sees the
	// flag and returns early must find the skip already cleared
	runtime.XgoSetTrapSkip(false)
	atomic.StoreInt32(&trapEnabled, 1)
}

// enableTrapFor makes the function of funcInfo call trap again, so that
// registering a mock for one function does not slow down every other
// instrumented call:
//   - functions guarded by xgo carry their own flag (as FuncInfo.Var, which
//     only variables use otherwise), and only that flag is set
//   - functions without a flag are not guarded (dependencies instrumented by
//     the compiler, the trace package) and always call trap: only the
//     guard reads the skip, XgoTrap itself does not
//   - variables are trapped on every access, see markGlobalInterceptorHolderUsed
//   - when the function is unknown (nil, e.g. a generic instantiation not
//     registered by PC), every instrumented function is enabled
func enableTrapFor(funcInfo *core.FuncInfo) {
	if funcInfo == nil {
		enableTrap()
		return
	}
	if funcInfo.Kind != core.Kind_Func {
		return
	}
	if flag, ok := funcInfo.Var.(*uint32); ok && flag != nil {
		atomic.StoreUint32(flag, 1)
	}
}

// enableTrapForPC is enableTrapFor the function registered at pc.
func enableTrapForPC(pc uintptr) {
	enableTrapFor(functab.InfoPC(pc))
}

// initTrapSkip lets instrumented functions skip the trap until the first
// registration, unless one already happened.
func initTrapSkip() {
	trapSkipMu.Lock()
	defer trapSkipMu.Unlock()
	if atomic.LoadInt32(&trapEnabled) == 0 {
		runtime.XgoSetTrapSkip(true)
	}
}

type StackData struct {
	hasStartedTracing bool

	filterTrace     func(funcInfo *core.FuncInfo) bool
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
	var mockList []*mockHolder
	if c != nil {
		mockList = c.interceptors.mock[pc]
	}
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
	var mockList []*varMockHolder
	if c != nil {
		mockList = c.interceptors.varMock[varAddr]
	}
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
	var mockList []*varMockHolder
	if c != nil {
		mockList = c.interceptors.varPtrMock[varAddr]
	}
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
	var localRecorders []*recorderHolder
	if c != nil {
		localRecorders = c.interceptors.recorder[pc]
	}
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
	var localVarRecorders []*varRecordHolder
	if c != nil {
		localVarRecorders = c.interceptors.varRecorder[varAddr]
	}
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
	var localVarPtrRecorders []*varRecordHolder
	if c != nil {
		localVarPtrRecorders = c.interceptors.varPtrRecorder[varAddr]
	}
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
	var localInterceptors []*recorderHolder
	if c != nil {
		localInterceptors = c.interceptors.interceptors
	}
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
