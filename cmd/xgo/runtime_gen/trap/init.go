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
}

func associateG(curG *_G, newG *_G) {
	stack := curG.GetStack()
	if stack == nil {
		return
	}

	// inherit mock
	newStack := newG.GetOrAttachStack()
	newStack.mock = cloneMocks(stack.mock)

	// associate trace
	if stack.hasStartedTracing {
		newStack.hasStartedTracing = true

		if stack.Top != nil {
			child := &StackEntry{
				StartNs:  time.Since(stack.Begin).Nanoseconds(),
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

func cloneMocks(mock map[uintptr][]*mockHolder) map[uintptr][]*mockHolder {
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
