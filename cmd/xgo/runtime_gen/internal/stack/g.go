package stack

import (
	"time"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

// G points to runtime.G
type G uintptr

func GetG() G {
	return G(unsafe.Pointer(runtime.GetG()))
}

func (g G) DetachStack() {
	runtime.AsG(unsafe.Pointer(g)).Delete(gStackKey)
}

func (g G) AttachStack(stack *Stack) {
	if stack == nil {
		panic("requires stack")
	}
	prevStack := runtime.AsG(unsafe.Pointer(g)).Get(gStackKey)
	if prevStack != nil {
		panic("stack already attached")
	}

	runtime.AsG(unsafe.Pointer(g)).Set(gStackKey, stack)
}

func (g G) GetStack() *Stack {
	stack := runtime.AsG(unsafe.Pointer(g)).Get(gStackKey)
	if stack == nil {
		return nil
	}
	return stack.(*Stack)
}

func (g G) GetOrAttachStack() *Stack {
	prevStack := runtime.AsG(unsafe.Pointer(g)).Get(gStackKey)
	if prevStack != nil {
		return prevStack.(*Stack)
	}
	stack := &Stack{
		Begin: time.Now(),
	}
	runtime.AsG(unsafe.Pointer(g)).Set(gStackKey, stack)
	return stack
}
