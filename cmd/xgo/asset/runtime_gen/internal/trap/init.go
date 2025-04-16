package trap

import (
	"unsafe"

	"github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/internal/stack"
)

func init() {
	runtime.XgoSetTrap(trap)
	runtime.XgoSetVarTrap(trapVar)
	runtime.XgoSetVarPtrTrap(trapVarPtr)
	runtime.XgoOnCreateG(func(g unsafe.Pointer, childG unsafe.Pointer) {
		inerhitStack(stack.G(g), stack.G(childG))
	})
	runtime.XgoOnExitG(func() {
		g := stack.GetG()
		stack := g.GetStack()
		if stack == nil {
			return
		}
		if stack.End.IsZero() {
			// fill end
			stack.End = runtime.XgoRealTimeNow()
		}
	})
}

func inerhitStack(curG stack.G, newG stack.G) {
	curStack := curG.GetStack()
	if curStack == nil {
		return
	}

	newStack := newG.GetOrAttachStack()
	newStackData := getOrAttachStackDataOf(newStack)

	stackData := getStackDataOf(curStack)

	newInterceptors := &newStackData.interceptors
	existingInterceptors := &stackData.interceptors

	// inherit mock
	newInterceptors.mock = cloneFuncMocks(existingInterceptors.mock)
	newInterceptors.varMock = cloneVarMocks(existingInterceptors.varMock)
	newInterceptors.varPtrMock = cloneVarMocks(existingInterceptors.varPtrMock)

	// recorder
	newInterceptors.recorder = cloneFuncRecordMapping(existingInterceptors.recorder)
	newInterceptors.varRecorder = cloneVarRecordMapping(existingInterceptors.varRecorder)
	newInterceptors.varPtrRecorder = cloneVarRecordMapping(existingInterceptors.varPtrRecorder)

	// interceptors
	newInterceptors.interceptors = cloneRecorderList(existingInterceptors.interceptors)

	// associate trace
	if stackData.hasStartedTracing {
		newStackData.hasStartedTracing = true

		if curStack.Top != nil {
			child := &stack.Entry{
				BeginNs:  newStack.Begin.Sub(curStack.Begin).Nanoseconds(),
				Go:       true,
				FuncName: "go",
				GetStack: func() *stack.Stack {
					return newStack
				},
			}
			curStack.Top.Children = append(curStack.Top.Children, child)
		}
	}
}

func cloneFuncMocks(mock map[uintptr][]*mockHolder) map[uintptr][]*mockHolder {
	if mock == nil {
		return nil
	}
	newMock := make(map[uintptr][]*mockHolder, len(mock))
	for pc, mocks := range mock {
		newMocks := make([]*mockHolder, len(mocks))
		for i, m := range mocks {
			newMocks[i] = &mockHolder{
				wantRecvPtr: m.wantRecvPtr,
				mock:        m.mock,
			}
		}
		newMock[pc] = newMocks
	}
	return newMock
}

func cloneVarMocks(mock map[uintptr][]*varMockHolder) map[uintptr][]*varMockHolder {
	if mock == nil {
		return nil
	}
	newMock := make(map[uintptr][]*varMockHolder, len(mock))
	for pc, mocks := range mock {
		newMocks := make([]*varMockHolder, len(mocks))
		for i, m := range mocks {
			newMocks[i] = &varMockHolder{
				mock: m.mock,
			}
		}
		newMock[pc] = newMocks
	}
	return newMock
}

func cloneFuncRecordMapping(recorders map[uintptr][]*recorderHolder) map[uintptr][]*recorderHolder {
	if recorders == nil {
		return nil
	}
	newRecorders := make(map[uintptr][]*recorderHolder, len(recorders))
	for pc, recorders := range recorders {
		newRecorders[pc] = make([]*recorderHolder, len(recorders))
		for i, recorder := range recorders {
			newRecorders[pc][i] = &recorderHolder{
				pre:  recorder.pre,
				post: recorder.post,
			}
		}
	}
	return newRecorders
}

func cloneVarRecordMapping(recorders map[uintptr][]*varRecordHolder) map[uintptr][]*varRecordHolder {
	if recorders == nil {
		return nil
	}
	newRecorders := make(map[uintptr][]*varRecordHolder, len(recorders))
	for pc, recorders := range recorders {
		newRecorders[pc] = make([]*varRecordHolder, len(recorders))
		for i, recorder := range recorders {
			newRecorders[pc][i] = &varRecordHolder{
				pre:  recorder.pre,
				post: recorder.post,
			}
		}
	}
	return newRecorders
}

func cloneRecorderList(recorders []*recorderHolder) []*recorderHolder {
	if recorders == nil {
		return nil
	}
	newRecorders := make([]*recorderHolder, len(recorders))
	for i, recorder := range recorders {
		newRecorders[i] = &recorderHolder{
			pre:  recorder.pre,
			post: recorder.post,
		}
	}
	return newRecorders
}
