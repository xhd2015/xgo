package trap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

const __XGO_SKIP_TRAP = true

var ErrAbort error = errors.New("abort trap interceptor")

// link by compiler
func __xgo_link_getcurg() unsafe.Pointer {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_getcurg.(xgo required)")
	return nil
}

func __xgo_link_init_finished() bool {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_init_finished.(xgo required)")
	return false
}

func __xgo_link_on_goexit(fn func()) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_goexit.(xgo required)")
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
var localInterceptors sync.Map // goroutine ptr -> *interceptorList

func AddInterceptor(interceptor *Interceptor) func() {
	ensureTrapInstall()
	if __xgo_link_init_finished() {
		return addLocalInterceptor(interceptor)
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
	dispose := addLocalInterceptor(interceptor)
	defer dispose()
	f()
}

const methodSuffix = "-fm"

// Inspect make a call to f to capture its receiver pointer if is
// is bound method
// It can be used to get the unwrapped innermost function of a method
// wrapper.
func Inspect(f interface{}) (recvPtr interface{}, funcInfo *core.FuncInfo) {
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func {
		panic(fmt.Errorf("Inspect requires func, given: %s", fn.Kind().String()))
	}
	pc := fn.Pointer()

	// funcs, closures and type functions can be found directly by PC
	funcInfo = functab.InfoPC(pc)
	if funcInfo != nil {
		return nil, funcInfo
	}

	fullName := __xgo_link_get_pc_name(pc)

	var isMethod bool
	if strings.HasSuffix(fullName, methodSuffix) {
		isMethod = true
		fullName = fullName[:len(fullName)-len(methodSuffix)]
	}

	funcInfo = functab.GetFuncByFullName(fullName)
	if funcInfo == nil {
		return nil, nil
	}
	// plain function
	if !isMethod {
		return nil, funcInfo
	}

	WithInterceptor(&Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			funcInfo = f
			recvPtr = args.GetFieldIndex(0).Ptr()
			return nil, ErrAbort
		},
	}, func() {
		fnType := fn.Type()
		nargs := fnType.NumIn()
		args := make([]reflect.Value, nargs)
		for i := 0; i < nargs; i++ {
			args[i] = reflect.New(fnType.In(i)).Elem()
		}
		if !fnType.IsVariadic() {
			fn.Call(args)
		} else {
			fn.CallSlice(args)
		}
	})
	return recvPtr, funcInfo
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

func ClearLocalInterceptors() {
	clearLocalInterceptorsAndMark()
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
func addLocalInterceptor(interceptor *Interceptor) func() {
	ensureTrapInstall()
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

func clearLocalInterceptorsAndMark() {
	key := __xgo_link_getcurg()
	localInterceptors.Delete(key)

	clearTrappingMark()
}
