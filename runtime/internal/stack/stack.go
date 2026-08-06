package stack

import (
	"sync"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
)

// Stack is a per-goroutine call tree. All tree mutations and export snapshots
// go through mu so concurrent Push (child) vs Export (parent) is race-free
// under -race (see race_export / issue #341 follow-up Phase 1).
//
// IMPORTANT: never call instrumented functions (e.g. time.Time methods that
// may trap) while holding mu — trap re-enters GetData and deadlocks.
type Stack struct {
	mu sync.Mutex

	// Begin is fixed after stack creation; safe to read without mu for
	// computing relative nanoseconds outside the lock.
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

	BeginNs  int64
	EndNs    int64
	Finished bool

	Children []*Entry
	Data     EntryData

	Go bool // has go keyword
	// only valid when Go==true
	GetStack func() *Stack

	FuncName string
	File     string
	Line     int

	HitMock   bool
	Panic     bool
	PanicLine int
	Error     string

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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.pushLocked(cur)
}

func (c *Stack) pushLocked(cur *Entry) *Entry {
	// Historical behavior: Push bumps MaxEntryID even though Entry.ID was
	// already assigned in NewEntry.
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
	// Compute relative time outside the lock (time methods may trap).
	beginNs := begin.UnixNano() - c.Begin.UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.newEntryLocked(beginNs, fnName)
}

func (c *Stack) newEntryLocked(beginNs int64, fnName string) *Entry {
	c.MaxEntryID++
	return &Entry{
		ID:       c.MaxEntryID,
		FuncName: fnName,
		BeginNs:  beginNs,
	}
}

// PushNew creates an entry, pushes it, runs setup, and increments Depth under
// one lock so the tree is never observed half-updated by Export.
// MaxEntryID is bumped twice (newEntry + push), matching historical NewEntry+Push.
func (c *Stack) PushNew(begin time.Time, fnName string, setup func(*Entry)) (cur *Entry, oldTop *Entry) {
	beginNs := begin.UnixNano() - c.Begin.UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()
	cur = c.newEntryLocked(beginNs, fnName)
	oldTop = c.pushLocked(cur)
	if setup != nil {
		setup(cur)
	}
	c.Depth++
	return cur, oldTop
}

// Finish completes an entry, restores Top, decrements Depth under lock.
// finish must not call instrumented functions (no time.Time methods, no traps).
func (c *Stack) Finish(cur *Entry, oldTop *Entry, end time.Time, setStackEnd bool, finish func(*Entry)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if setStackEnd {
		c.End = end
	}
	if finish != nil {
		finish(cur)
	}
	c.Top = oldTop
	c.Depth--
}

// AppendGoChild attaches a synthetic "go" node under the current Top under lock.
func (c *Stack) AppendGoChild(beginNs int64, getStack func() *Stack) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Top == nil {
		return
	}
	child := &Entry{
		BeginNs:  beginNs,
		Go:       true,
		FuncName: "go",
		GetStack: getStack,
	}
	c.Top.Children = append(c.Top.Children, child)
}

// SetEndIfZero records stack end time once (goroutine exit).
func (c *Stack) SetEndIfZero(t time.Time) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	// Do not call time.Time methods (e.g. IsZero) while holding mu — they are
	// instrumented and re-enter trap → GetData → same mutex (deadlock).
	if c.End == (time.Time{}) {
		c.End = t
	}
}

// Snapshot copies the tree for race-free export. GetStack funcs still point at
// the live Stack; callers Export those under their own locks.
// Must not invoke instrumented code while holding mu.
func (c *Stack) Snapshot() (begin, end time.Time, roots []*Entry) {
	if c == nil {
		return time.Time{}, time.Time{}, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Begin, c.End, cloneEntries(c.Roots)
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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Data[key]
}

func (c *Stack) SetData(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Data == nil {
		c.Data = make(map[interface{}]interface{}, 1)
	}
	c.Data[key] = value
}

func cloneEntries(entries []*Entry) []*Entry {
	if entries == nil {
		return nil
	}
	out := make([]*Entry, len(entries))
	for i, e := range entries {
		out[i] = cloneEntry(e)
	}
	return out
}

func cloneEntry(e *Entry) *Entry {
	if e == nil {
		return nil
	}
	cp := *e
	cp.Children = cloneEntries(e.Children)
	if e.Data != nil {
		cp.Data = make(EntryData, len(e.Data))
		for k, v := range e.Data {
			cp.Data[k] = v
		}
	}
	return &cp
}
