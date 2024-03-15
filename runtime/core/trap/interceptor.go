package trap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core/functab"
)

const __XGO_SKIP_TRAP = true

var ErrAbort error = errors.New("abort trap interceptor")

type FuncArgs struct {
	Recv    interface{}
	Args    []interface{}
	Results []interface{}
}

type Interceptor struct {
	Pre  func(ctx context.Context, f *functab.FuncInfo, args *FuncArgs) (data interface{}, err error)
	Post func(ctx context.Context, f *functab.FuncInfo, args *FuncArgs, data interface{}) error
}

var interceptors []*Interceptor
var localInterceptors sync.Map // goroutine ptr -> *interceptorList

func AddInterceptor(interceptor *Interceptor) {
	ensureSetupTrap()
	interceptors = append(interceptors, interceptor)
}

func GetInterceptors() []*Interceptor {
	return interceptors
}
func GetLocalInterceptors() []*Interceptor {
	key := __xgo_link_getcurg()
	val, ok := localInterceptors.Load(key)
	if !ok {
		return nil
	}
	return val.(*interceptorList).interceptors
}

func WithLocalInterceptor(interceptor *Interceptor, f func()) {
	dispose := AddLocalInterceptor(interceptor)
	defer dispose()
	f()
}

func GetAllInterceptors() []*Interceptor {
	locals := GetLocalInterceptors()
	global := GetInterceptors()
	if len(locals) == 0 {
		return global
	}
	if len(global) == 0 {
		return locals
	}
	// run locals first(in reversed order)
	return append(global, locals...)
}

// returns a function to dispose the key
// NOTE: if not called correctly,there might be memory leak
func AddLocalInterceptor(interceptor *Interceptor) func() {
	ensureSetupTrap()
	key := __xgo_link_getcurg()
	list := &interceptorList{}
	val, loaded := localInterceptors.LoadOrStore(key, list)
	if loaded {
		list = val.(*interceptorList)
	}
	list.interceptors = append(list.interceptors, interceptor)

	removed := false
	// used to remove the local interceptor
	return func() {
		if removed {
			panic(fmt.Errorf("remove interceptor more than once"))
		}
		removed = true
		var idx int = -1
		for i, intc := range list.interceptors {
			if intc == interceptor {
				idx = i
				break
			}
		}
		if idx < 0 {
			panic(fmt.Errorf("interceptor leaked"))
		}
		n := len(list.interceptors)
		for i := idx; i < n-1; i++ {
			list.interceptors[i] = list.interceptors[i+1]
		}
		list.interceptors = list.interceptors[:n-1]
		if len(list.interceptors) == 0 {
			// remove the entry from map to prevent memory leak
			localInterceptors.Delete(key)
		}
	}
}

type interceptorList struct {
	interceptors []*Interceptor
}

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	panic(errors.New("xgo failed to link __xgo_link_getcurg"))
}
