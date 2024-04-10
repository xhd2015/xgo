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

var interceptors []*Interceptor
var localInterceptors sync.Map // goroutine ptr -> *interceptorGroup

func AddInterceptor(interceptor *Interceptor) func() {
	ensureTrapInstall()
	if __xgo_link_init_finished() {
		dispose, _ := addLocalInterceptor(interceptor, false)
		return dispose
	}
	interceptors = append(interceptors, interceptor)
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
	dispose, _ := addLocalInterceptor(interceptor, false)
	defer dispose()
	f()
}

// WithOverride override local and global interceptors
// in current goroutine temporarily, it returns a function
// that can be used to cancel the override.
func WithOverride(interceptor *Interceptor, f func()) {
	_, disposeGroup := addLocalInterceptor(interceptor, true)
	defer disposeGroup()
	f()
}

func GetInterceptors() []*Interceptor {
	return interceptors
}

func GetLocalInterceptors() []*Interceptor {
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
	var locals []*Interceptor
	var g int
	if group != nil {
		gi := group.currentGroupInterceptors()
		if gi != nil {
			g = group.currentGroup()
			if gi.override {
				return gi.list, g
			}
			locals = gi.list
		}
	}
	global := GetInterceptors()
	if len(locals) == 0 {
		return global, g
	}
	if len(global) == 0 {
		return locals, g
	}
	// run locals first(in reversed order)
	return append(global[:len(global):len(global)], locals...), g
}

// returns a function to dispose the key
// NOTE: if not called correctly,there might be memory leak
func addLocalInterceptor(interceptor *Interceptor, override bool) (removeInterceptor func(), removeGroup func()) {
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
	list.appendToCurrentGroup(interceptor)

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
		interceptors := list.groups[g].list
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
		list.groups[g].list = interceptors
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
	list     []*Interceptor
}

func (c *interceptorList) append(interceptor *Interceptor) {
	c.list = append(c.list, interceptor)
}

func (c *interceptorGroup) appendToCurrentGroup(interceptor *Interceptor) {
	g := c.currentGroup()
	c.groups[g].append(interceptor)
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
		list:     make([]*Interceptor, 0, 1),
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
