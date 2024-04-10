package trap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
)

const __XGO_SKIP_TRAP = true

var ErrAbort error = errors.New("abort trap interceptor")

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_getcurg(requires xgo).")
	return nil
}

func __xgo_link_init_finished() bool {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_init_finished(requires xgo).")
	return false
}

func __xgo_link_on_goexit(fn func()) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_goexit(requires xgo).")
}
func __xgo_link_get_pc_name(pc uintptr) string {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_get_pc_name(requires xgo)")
	return ""
}

func init() {
	__xgo_link_on_goexit(clearLocalInterceptorsAndMark)
}

type Interceptor struct {
	Pre  func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object) (data interface{}, err error)
	Post func(ctx context.Context, f *core.FuncInfo, args core.Object, result core.Object, data interface{}) error
}

type interceptorManager struct {
	head []*Interceptor
	tail []*Interceptor
}

func (c *interceptorManager) copy() *interceptorManager {
	if c == nil {
		return nil
	}
	head := make([]*Interceptor, len(c.head))
	tail := make([]*Interceptor, len(c.tail))
	copy(head, c.head)
	copy(tail, c.tail)
	return &interceptorManager{
		head: head,
		tail: tail,
	}
}

var interceptors = &interceptorManager{}
var localInterceptors sync.Map // goroutine ptr -> *interceptorGroup

func AddInterceptor(interceptor *Interceptor) func() {
	return addInterceptor(interceptor, false)
}

func AddInterceptorHead(interceptor *Interceptor) func() {
	return addInterceptor(interceptor, true)
}
func addInterceptor(interceptor *Interceptor, head bool) func() {
	ensureTrapInstall()
	if __xgo_link_init_finished() {
		dispose, _ := addLocalInterceptor(interceptor, false, head)
		return dispose
	}
	if head {
		interceptors.head = append(interceptors.head, interceptor)
	} else {
		interceptors.tail = append(interceptors.tail, interceptor)
	}
	return func() {
		panic("global interceptor cannot be cancelled, if you want to cancel a global interceptor, use WithInterceptor")
	}
}

// WithInterceptor executes given f with interceptor
// setup. It can be used from init phase safely.
// it clears the interceptor after f finishes.
// the interceptor will be added to head, so it
// will gets firstly invoked.
// f cannot be nil.
//
// NOTE: the implementation uses addLocalInterceptor
// even from init because it will be soon cleared
// without causing concurrent issues.
func WithInterceptor(interceptor *Interceptor, f func()) {
	dispose, _ := addLocalInterceptor(interceptor, false, false)
	defer dispose()
	f()
}

// WithOverride override local and global interceptors
// in current goroutine temporarily, it returns a function
// that can be used to cancel the override.
func WithOverride(interceptor *Interceptor, f func()) {
	_, disposeGroup := addLocalInterceptor(interceptor, true, false)
	defer disposeGroup()
	f()
}

func GetInterceptors() []*Interceptor {
	return interceptors.getInterceptors()
}

func (c *interceptorManager) getInterceptors() []*Interceptor {
	return mergeInterceptors(c.tail, c.head)
}

func mergeInterceptors(groups ...[]*Interceptor) []*Interceptor {
	n := 0
	for _, g := range groups {
		n += len(g)
	}
	list := make([]*Interceptor, 0, n)
	for _, g := range groups {
		list = append(list, g...)
	}
	return list
}

func GetLocalInterceptors() []*Interceptor {
	g := getLocalInterceptorList()
	if g == nil {
		return nil
	}
	return g.getInterceptors()
}
func getLocalInterceptorList() *interceptorManager {
	group := getLocalInterceptorGroup()
	if group == nil {
		return nil
	}
	gi := group.currentGroupInterceptors()
	if gi == nil {
		return nil
	}
	return gi.list
}
func getLocalInterceptorGroup() *interceptorGroup {
	key := uintptr(__xgo_link_getcurg())
	val, ok := localInterceptors.Load(key)
	if !ok {
		return nil
	}
	return val.(*interceptorGroup)
}

func ClearLocalInterceptors() {
	clearLocalInterceptorsAndMark()
}

func GetAllInterceptors() []*Interceptor {
	res, _ := getAllInterceptors()
	return res
}

func getAllInterceptors() ([]*Interceptor, int) {
	group := getLocalInterceptorGroup()
	var localHead []*Interceptor
	var localTail []*Interceptor
	var g int
	if group != nil {
		gi := group.currentGroupInterceptors()
		if gi != nil {
			g = group.currentGroup()
			if gi.override {
				return gi.list.getInterceptors(), g
			}
			localHead = gi.list.head
			localTail = gi.list.tail
		}
	}
	globalHead := interceptors.head
	globalTail := interceptors.tail

	// run locals first(in reversed order)
	return mergeInterceptors(globalTail, localTail, globalHead, localHead), g
}

// returns a function to dispose the key
// NOTE: if not called correctly,there might be memory leak
func addLocalInterceptor(interceptor *Interceptor, override bool, head bool) (removeInterceptor func(), removeGroup func()) {
	ensureTrapInstall()
	key := uintptr(__xgo_link_getcurg())
	list := &interceptorGroup{}
	val, loaded := localInterceptors.LoadOrStore(key, list)
	if loaded {
		list = val.(*interceptorGroup)
	}
	// ensure at least one group
	if override || list.groupsEmpty() {
		list.enterNewGroup(override)
	}
	g := list.currentGroup()
	list.appendToCurrentGroup(interceptor, head)

	removedInterceptor := false
	// used to remove the local interceptor
	removeInterceptor = func() {
		if removedInterceptor {
			panic(fmt.Errorf("remove interceptor more than once"))
		}
		curKey := uintptr(__xgo_link_getcurg())
		if key != curKey {
			panic(fmt.Errorf("remove interceptor from another goroutine"))
		}
		curG := list.currentGroup()
		if curG != g {
			panic(fmt.Errorf("interceptor group changed: previous=%d, current=%d", g, curG))
		}
		manager := list.groups[g].list

		interceptors := manager.tail
		if head {
			interceptors = manager.head
		}
		removedInterceptor = true
		var idx int = -1
		for i, intc := range interceptors {
			if intc == interceptor {
				idx = i
				break
			}
		}
		if idx < 0 {
			panic(fmt.Errorf("interceptor leaked"))
		}
		n := len(interceptors)
		for i := idx; i < n-1; i++ {
			interceptors[i] = interceptors[i+1]
		}
		interceptors = interceptors[:n-1]
		if head {
			manager.head = interceptors
		} else {
			manager.tail = interceptors
		}
	}

	removedGroup := false
	removeGroup = func() {
		if removedGroup {
			panic(fmt.Errorf("remove group more than once"))
		}
		curKey := uintptr(__xgo_link_getcurg())
		if key != curKey {
			panic(fmt.Errorf("remove group from another goroutine"))
		}
		curG := list.currentGroup()
		if curG != g {
			panic(fmt.Errorf("interceptor group changed: previous=%d, current=%d", g, curG))
		}
		list.exitGroup()
		removedInterceptor = true
	}

	return removeInterceptor, removeGroup
}

type interceptorGroup struct {
	groups []*interceptorList
}

type interceptorList struct {
	override bool
	list     *interceptorManager
}

func (c *interceptorList) append(interceptor *Interceptor, head bool) {
	if head {
		c.list.head = append(c.list.head, interceptor)
	} else {
		c.list.tail = append(c.list.tail, interceptor)
	}
}

func (c *interceptorGroup) appendToCurrentGroup(interceptor *Interceptor, head bool) {
	g := c.currentGroup()
	c.groups[g].append(interceptor, head)
}

func (c *interceptorGroup) groupsEmpty() bool {
	return len(c.groups) == 0
}
func (c *interceptorGroup) currentGroup() int {
	n := len(c.groups)
	return n - 1
}
func (c *interceptorGroup) currentGroupInterceptors() *interceptorList {
	g := c.currentGroup()
	if g < 0 {
		return nil
	}
	return c.groups[g]
}

func (c *interceptorGroup) enterNewGroup(override bool) {
	c.groups = append(c.groups, &interceptorList{
		override: override,
		list:     &interceptorManager{},
	})
}
func (c *interceptorGroup) exitGroup() {
	n := len(c.groups)
	if n == 0 {
		panic("exit no group")
	}
	c.groups = c.groups[:n-1]
}

func clearLocalInterceptorsAndMark() {
	key := uintptr(__xgo_link_getcurg())
	localInterceptors.Delete(key)
	bypassMapping.Delete(key)

	clearTrappingMarkAllGroup()
}
