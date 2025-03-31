package trap

import (
	"time"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

func init() {
	runtime.XgoSetTrap(trap)
	runtime.XgoSetVarTrap(trapVar)
	runtime.XgoSetVarPtrTrap(trapVarPtr)
	runtime.XgoOnCreateG(func(g unsafe.Pointer, childG unsafe.Pointer) {
		associateG((*_G)(g), (*_G)(childG))
	})
	runtime.XgoOnExitG(func() {
		g := _GetG()
		stack := g.GetStack()
		if stack == nil {
			return
		}
		if stack.End.IsZero() {
			// fill end
			stack.End = time.Now()
		}
	})
}

func associateG(curG *_G, newG *_G) {
	stack := curG.GetStack()
	if stack == nil {
		return
	}

	// inherit mock
	newStack := newG.GetOrAttachStack()
	newStack.mock = cloneFuncMocks(stack.mock)
	newStack.varMock = cloneVarMocks(stack.varMock)
	newStack.varPtrMock = cloneVarMocks(stack.varPtrMock)

	// associate trace
	if stack.hasStartedTracing {
		newStack.hasStartedTracing = true

		if stack.Top != nil {
			child := &StackEntry{
				BeginNs:  newStack.Begin.Sub(stack.Begin).Nanoseconds(),
				Go:       true,
				FuncName: "go",
				GetStack: func() *Stack {
					return newStack
				},
			}
			stack.Top.Children = append(stack.Top.Children, child)
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
