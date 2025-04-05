package stack

import (
	"unsafe"

	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

// NilGStack is a guard pointer for when g is nil
// because m is just starting on g0
var NilGStack = &Stack{}

const NilG = G(uintptr(0)) // nil

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
	if g == NilG {
		return NilGStack
	}
	stack := runtime.AsG(unsafe.Pointer(g)).Get(gStackKey)
	if stack == nil {
		return nil
	}
	return stack.(*Stack)
}

func (g G) GetOrAttachStack() *Stack {
	if g == NilG {
		panic("cannot attach stack on nil g(m might just be starting)")
	}
	prevStack := runtime.AsG(unsafe.Pointer(g)).Get(gStackKey)
	if prevStack != nil {
		return prevStack.(*Stack)
	}
	stack := &Stack{
		Begin: runtime.XgoRealTimeNow(),
	}
	runtime.AsG(unsafe.Pointer(g)).Set(gStackKey, stack)
	return stack
}
