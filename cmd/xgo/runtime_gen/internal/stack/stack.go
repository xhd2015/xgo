package stack

import (
	"time"

	"github.com/xhd2015/xgo/runtime/core"
)

type Stack struct {
	Begin      time.Time
	End        time.Time
	MaxEntryID int

	Roots []*Entry
	Top   *Entry
	Depth int

	Data map[interface{}]interface{}
}

func Get() *Stack {
	return GetG().GetStack()
}

func GetOrAttach() *Stack {
	return GetG().GetOrAttachStack()
}

type EntryData map[interface{}]interface{}

type Entry struct {
	ID       int
	ParentID int

	FuncInfo *core.FuncInfo

	BeginNs int64
	EndNs   int64

	Children []*Entry
	Data     EntryData

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

type gStackKeyType struct{}

var gStackKey = gStackKeyType{}

// Attach attaches a stack for recording
func Attach(stack *Stack) {
	GetG().AttachStack(stack)
}

func Detach() {
	GetG().DetachStack()
}

// push returns the old top
func (c *Stack) Push(cur *Entry) *Entry {
	c.MaxEntryID++
	oldTop := c.Top
	if oldTop == nil {
		c.Roots = append(c.Roots, cur)
	} else {
		cur.ParentID = oldTop.ID
		oldTop.Children = append(oldTop.Children, cur)
	}
	c.Top = cur
	return oldTop
}

func (c *Stack) NewEntry(begin time.Time, fnName string) *Entry {
	c.MaxEntryID++
	cur := &Entry{
		ID:       c.MaxEntryID,
		FuncName: fnName,
		BeginNs:  begin.UnixNano() - c.Begin.UnixNano(),
	}
	return cur
}

type Field struct {
	Name  string
	Value interface{}
}

func (s *Entry) SetData(key, value interface{}) {
	if s.Data == nil {
		s.Data = make(EntryData)
	}
	s.Data[key] = value
}

func (s *Entry) GetData(key interface{}) interface{} {
	return s.Data[key]
}

type StackArgs []interface{}

func (c *Stack) GetData(key interface{}) interface{} {
	return c.Data[key]
}

func (c *Stack) SetData(key, value interface{}) {
	if c.Data == nil {
		c.Data = make(map[interface{}]interface{}, 1)
	}
	c.Data[key] = value
}
