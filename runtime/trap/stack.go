package trap

import (
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap/stack_model"
)

type Stack struct {
	Begin time.Time
	End   time.Time
	MaxID int

	hasStartedTracing bool

	onFinish   func(stack stack_model.IStack)
	OutputFile string

	Roots []*StackEntry
	Top   *StackEntry
	Depth int

	Data map[interface{}]interface{}

	// pc->mock
	mock       map[uintptr][]*mockHolder
	varMock    map[uintptr][]*varMockHolder
	varPtrMock map[uintptr][]*varMockHolder

	// pc->recorder
	recorder       map[uintptr][]*recorderHolder
	varRecorder    map[uintptr][]*varRecordHolder
	varPtrRecorder map[uintptr][]*varRecordHolder

	inspecting func(pc uintptr, recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{})
}

type StackEntryData map[interface{}]interface{}

type StackEntry struct {
	ID       int
	ParentID int

	FuncInfo *core.FuncInfo

	BeginNs int64
	EndNs   int64

	Children []*StackEntry
	Data     StackEntryData

	Go bool // has go keyword
	// only valid when Go==true
	GetStack func() *Stack

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
	_GetG().AttachStack(stack)
}

func (g *_G) AttachStack(stack *Stack) {
	if stack == nil {
		panic("requires stack")
	}
	prevStack := g.Get(stackKey)
	if prevStack != nil {
		panic("stack already attached")
	}

	g.Set(stackKey, stack)
}

func GetStack() *Stack {
	g := _GetG()
	return g.GetStack()
}

func (g *_G) GetStack() *Stack {
	stack := g.Get(stackKey)
	if stack == nil {
		return nil
	}
	return stack.(*Stack)
}

func GetOrAttachStack() *Stack {
	return _GetG().GetOrAttachStack()
}

func (g *_G) GetOrAttachStack() *Stack {
	prevStack := g.Get(stackKey)
	if prevStack != nil {
		return prevStack.(*Stack)
	}
	stack := &Stack{
		Begin: time.Now(),
	}
	g.Set(stackKey, stack)
	return stack
}

func DetachStack() {
	_GetG().DetachStack()
}

func (g *_G) DetachStack() {
	g.Delete(stackKey)
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
