package trace_runtime

import "time"

type Stack struct {
	Begin time.Time
	MaxID int

	OutputFile string

	Roots []*StackEntry
	Top   *StackEntry
	Depth int

	Data map[interface{}]interface{}

	OnEnter func(entry *StackEntry, pc uintptr, args StackArgs)
	OnExit  func(entry *StackEntry, pc uintptr, results []Field)
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
	Args     interface{}
	Results  interface{}
}

type stackKeyType struct{}

var stackKey = stackKeyType{}

// AttachStack attaches a stack for recording
func AttachStack(stack *Stack) {
	if stack == nil {
		panic("requires stack")
	}
	curg := GetG()
	st := curg.Get(stackKey)
	if st != nil {
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
