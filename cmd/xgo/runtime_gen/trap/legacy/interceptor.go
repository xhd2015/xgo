package legacy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/legacy"
)

var ErrAbort error = errors.New("abort trap interceptor")
var ErrSkip error = errors.New("skip trap interceptor")

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	if !legacy.V1_0_0 {
		return nil
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_getcurg(requires xgo).")
	return nil
}

func __xgo_link_init_finished() bool {
	if !legacy.V1_0_0 {
		return false
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_init_finished(requires xgo).")
	return false
}

func __xgo_link_on_goexit(fn func()) {
	if !legacy.V1_0_0 {
		return
	}
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_goexit(requires xgo).")
}
func __xgo_link_get_pc_name(pc uintptr) string {
	if !legacy.V1_0_0 {
		return ""
	}
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

var globalInterceptors = &interceptorManager{}
var localInterceptors sync.Map // goroutine ptr -> *interceptorGroup

// AddInterceptor add a general interceptor, disallowing re-entrant
func AddInterceptor(interceptor *Interceptor) func() {
	if !legacy.V1_0_0 {
		fmt.Fprintln(os.Stderr, "WARNING: AddFuncInterceptor no longer supported.")
		return nil
	}
	return addInterceptor(nil, interceptor, false)
}

func AddInterceptorHead(interceptor *Interceptor) func() {
	if !legacy.V1_0_0 {
		fmt.Fprintln(os.Stderr, "WARNING: AddFuncInterceptor no longer supported.")
		return nil
	}
	return addInterceptor(nil, interceptor, true)
}

// AddFuncInterceptor add func interceptor, allowing f to be re-entrant
func AddFuncInterceptor(f interface{}, interceptor *Interceptor) func() {
	if !legacy.V1_0_0 {
		fmt.Fprintln(os.Stderr, "WARNING: AddFuncInterceptor no longer supported.")
		return nil
	}
	_, fnInfo, pc, _ := InspectPC(f)
	if fnInfo == nil {
		panic(fmt.Errorf("failed to add func interceptor: %s", runtime.FuncForPC(pc).Name()))
	}
	return addInterceptor(fnInfo, interceptor, false)
}

func AddFuncInfoInterceptor(f *core.FuncInfo, interceptor *Interceptor) func() {
	if !legacy.V1_0_0 {
		fmt.Fprintln(os.Stderr, "WARNING: AddFuncInterceptor no longer supported.")
		return nil
	}
	if f == nil {
		panic(fmt.Errorf("func cannot be nil"))
	}
	return addInterceptor(f, interceptor, false)
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
	dispose, _ := addLocalInterceptor(nil, interceptor, false, false)
	defer dispose()
	f()
}

// WithOverride override local and global interceptors
// in current goroutine temporarily, it returns a function
// that can be used to cancel the override.
func WithOverride(interceptor *Interceptor, f func()) {
	_, disposeGroup := addLocalInterceptor(nil, interceptor, true, false)
	defer disposeGroup()
	f()
}
func WithFuncOverride(funcInfo *core.FuncInfo, interceptor *Interceptor, f func()) {
	_, disposeGroup := addLocalInterceptor(funcInfo, interceptor, true, false)
	defer disposeGroup()
	f()
}

func addInterceptor(f *core.FuncInfo, interceptor *Interceptor, head bool) func() {
	ensureTrapInstall()
	if __xgo_link_init_finished() {
		dispose, _ := addLocalInterceptor(f, interceptor, false, head)
		return dispose
	}
	Ignore(interceptor.Pre)
	Ignore(interceptor.Post)

	globalInterceptors.append(f, interceptor, false)
	return func() {
		if __xgo_link_init_finished() {
			// to ensure lock free
			panic("global interceptor cannot be cancelled after init, if you want to cancel a global interceptor, use WithInterceptor")
		}
		globalInterceptors.removeInterceptor(f, interceptor, false)
	}
}

type interceptorManager struct {
	head        []*Interceptor // always executed first
	tail        []*Interceptor
	funcMapping map[*core.FuncInfo][]*Interceptor // nested mapping
}

func (c *interceptorManager) hasAny() bool {
	if c == nil {
		return false
	}
	if len(c.head) > 0 {
		return true
	}
	if len(c.tail) > 0 {
		return true
	}
	for _, m := range c.funcMapping {
		if len(m) > 0 {
			return true
		}
	}
	return false
}

func (c *interceptorManager) copy() *interceptorManager {
	if c == nil {
		return nil
	}
	head := make([]*Interceptor, len(c.head))
	tail := make([]*Interceptor, len(c.tail))
	copy(head, c.head)
	copy(tail, c.tail)

	var funcMapping map[*core.FuncInfo][]*Interceptor
	if c.funcMapping != nil {
		funcMapping = make(map[*core.FuncInfo][]*Interceptor, len(c.funcMapping))
		for f, list := range c.funcMapping {
			cpList := make([]*Interceptor, len(list))
			copy(cpList, list)
			funcMapping[f] = cpList
		}
	}

	return &interceptorManager{
		head:        head,
		tail:        tail,
		funcMapping: funcMapping,
	}
}

func (c *interceptorManager) append(f *core.FuncInfo, interceptor *Interceptor, head bool) {
	if f != nil {
		if c.funcMapping == nil {
			c.funcMapping = make(map[*core.FuncInfo][]*Interceptor, 1)
		}
		c.funcMapping[f] = append(c.funcMapping[f], interceptor)
		return
	}
	if head {
		c.head = append(c.head, interceptor)
	} else {
		c.tail = append(c.tail, interceptor)
	}
}

func (c *interceptorManager) removeInterceptor(f *core.FuncInfo, interceptor *Interceptor, head bool) {
	if f == nil {
		if head {
			c.head = dropInterceptor(c.head, interceptor)
		} else {
			c.tail = dropInterceptor(c.tail, interceptor)
		}
	} else {
		newInterceptor := dropInterceptor(c.funcMapping[f], interceptor)
		if newInterceptor == nil {
			delete(c.funcMapping, f)
		} else {
			c.funcMapping[f] = newInterceptor
		}
	}
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

// f must not be nil
// if `noCommon` is set, only get f's mapping interceptors
// TODO: may allow trace when set `noLocalCommon`
func getAllInterceptors(f *core.FuncInfo, needCommon bool) ([]*Interceptor, int) {
	group := getLocalInterceptorGroup()

	var globalHead []*Interceptor
	var globalTail []*Interceptor

	var localHead []*Interceptor
	var localTail []*Interceptor
	var localFunc []*Interceptor

	var override bool
	var g int
	if group != nil {
		gi := group.currentGroupInterceptors()
		if gi != nil {
			g = group.currentGroup()
			override = gi.override
			if needCommon {
				localHead = gi.list.head
				localTail = gi.list.tail
			}
			localFunc = gi.list.funcMapping[f]
		}
	}

	if needCommon && !override {
		globalHead = globalInterceptors.head
		globalTail = globalInterceptors.tail
	}

	// run locals first(in reversed order)
	return mergeInterceptors(globalTail, localFunc, localTail, globalHead, localHead), g
}

// returns a function to dispose the key
// NOTE: if not called correctly,there might be memory leak
func addLocalInterceptor(f *core.FuncInfo, interceptor *Interceptor, override bool, head bool) (removeInterceptor func(), removeGroup func()) {
	ensureTrapInstall()
	Ignore(interceptor.Pre)
	Ignore(interceptor.Post)

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
	list.appendToCurrentGroup(f, interceptor, head)

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
		removedInterceptor = true
		list.groups[g].list.removeInterceptor(f, interceptor, head)
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

func dropInterceptor(interceptors []*Interceptor, interceptor *Interceptor) []*Interceptor {
	n := len(interceptors)
	idx := -1
	for i := 0; i < n; i++ {
		if interceptors[i] == interceptor {
			idx = i
			break
		}
	}
	if idx < 0 {
		panic("interceptor not found before removed")
	}

	for i := idx + 1; i < n; i++ {
		interceptors[i-1] = interceptors[i]
	}
	interceptors = interceptors[:n-1]
	if len(interceptors) == 0 {
		return nil
	}
	return interceptors
}

type interceptorGroup struct {
	groups []*interceptorList
}

type interceptorList struct {
	override bool
	list     *interceptorManager
}

func (c *interceptorGroup) appendToCurrentGroup(f *core.FuncInfo, interceptor *Interceptor, head bool) {
	g := c.currentGroup()
	c.groups[g].list.append(f, interceptor, head)
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
