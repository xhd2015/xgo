package trace_runtime

import (
	"time"
)

type Stack struct {
	Begin time.Time
	MaxID int

	hasStartedTracing bool
	OutputFile        string

	Roots []*StackEntry
	Top   *StackEntry
	Depth int

	Data map[interface{}]interface{}

	// pc->mock
	mock map[uintptr][]*mockHolder

	inspecting func(pc uintptr, recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{})
}

type StackEntryData map[interface{}]interface{}

type StackEntry struct {
	ID       int
	ParentID int

	StartNs int64
	EndNs   int64

	Children []*StackEntry
	Data     StackEntryData

	FuncName string
	File     string
	Line     int

	HitMock bool
	Panic   bool
	Error   string

	Args    interface{}
	Results interface{}
}

type stackKeyType struct{}

var stackKey = stackKeyType{}

// AttachStack attaches a stack for recording
func AttachStack(stack *Stack) {
	if stack == nil {
		panic("requires stack")
	}
	curg := GetG()
	prevStack := curg.Get(stackKey)
	if prevStack != nil {
		panic("stack already attached")
	}

	curg.Set(stackKey, stack)
}

func GetStack() *Stack {
	stack := GetG().Get(stackKey)
	if stack == nil {
		return nil
	}
	return stack.(*Stack)
}

func GetOrAttachStack() *Stack {
	g := GetG()
	prevStack := g.Get(stackKey)
	if prevStack != nil {
		return prevStack.(*Stack)
	}
	stack := &Stack{}
	g.Set(stackKey, stack)
	return stack
}

func DetachStack() {
	GetG().Delete(stackKey)
}

type Field struct {
	Name  string
	Value interface{}
}

func (s *StackEntry) SetData(key, value interface{}) {
	if s.Data == nil {
		s.Data = make(StackEntryData)
	}
	s.Data[key] = value
}

func (s *StackEntry) GetData(key interface{}) interface{} {
	return s.Data[key]
}

type StackArgs []interface{}

func (c *Stack) GetData(key interface{}) interface{} {
	return c.Data[key]
}
func (c *Stack) SetData(key, value interface{}) {
	c.Data[key] = value
}
