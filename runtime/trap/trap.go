package trap

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
)

var setupOnce sync.Once

func ensureTrapInstall() {
	setupOnce.Do(func() {
		__xgo_link_set_trap(trapFunc)
		__xgo_link_set_trap_var(trapVar)
	})
}
func init() {
	__xgo_link_on_gonewproc(func(g uintptr) {
		if isByPassing() {
			return
		}
		interceptors := GetLocalInterceptors()
		if len(interceptors) == 0 {
			return
		}
		copyInterceptors := make([]*Interceptor, len(interceptors))
		copy(copyInterceptors, interceptors)

		// inherit interceptors
		localInterceptors.Store(g, &interceptorList{
			interceptors: copyInterceptors,
		})
	})
}

func __xgo_link_set_trap(trapImpl func(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_set_trap(requires xgo).")
}

func __xgo_link_set_trap_var(trap func(pkgPath string, name string, tmpVarAddr interface{}, takeAddr bool)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_set_trap_var(requires xgo).")
}
func __xgo_link_on_gonewproc(f func(g uintptr)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_gonewproc(requires xgo).")
}

// Skip serves as mark to tell xgo not insert
// trap instructions for the function that
// calls Skip()
// NOTE: the function body is intenionally leave empty
// as trap.Skip() is just a mark that makes
// sense at compile time.
func Skip() {}

var trappingMark sync.Map // <goroutine key> -> struct{}{}
var trappingPC sync.Map   // <gorotuine key> -> PC

// link to runtime
// xgo:notrap
func trapFunc(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	interceptors := GetAllInterceptors()
	n := len(interceptors)
	if n == 0 {
		return nil, false
	}
	// setup context
	setTrappingPC(pc)
	defer clearTrappingPC()

	// NOTE: this may return nil for generic template
	var f *core.FuncInfo
	if !generic {
		f = functab.InfoPC(pc)
	}
	if f == nil {
		// generic and closure under go1.22
		f = functab.Info(pkgPath, identityName)
	}
	if f == nil {
		// no func found
		return nil, false
	}
	if f.RecvType != "" && methodHasBeenTrapped && recv == nil {
		// method
		// let go to the next interceptor
		return nil, false
	}
	return trap(f, interceptors, recv, args, results)
}

func trapVar(pkgPath string, name string, tmpVarAddr interface{}, takeAddr bool) {
	interceptors := GetAllInterceptors()
	n := len(interceptors)
	if n == 0 {
		return
	}
	identityName := name
	if takeAddr {
		identityName = "*" + name
	}
	fnInfo := functab.Info(pkgPath, identityName)
	if fnInfo == nil {
		return
	}
	if fnInfo.Kind != core.Kind_Var && fnInfo.Kind != core.Kind_VarPtr && fnInfo.Kind != core.Kind_Const {
		return
	}
	// NOTE: stop always ignored because this is a simple get
	post, _ := trap(fnInfo, interceptors, nil, nil, []interface{}{tmpVarAddr})
	if post != nil {
		// NOTE: must in defer, because in post we
		// may capture panic
		defer post()
	}
}
func trap(f *core.FuncInfo, interceptors []*Interceptor, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	if isByPassing() {
		return nil, false
	}
	dispose := setTrappingMark()
	if dispose == nil {
		return nil, false
	}
	defer dispose()

	// retrieve context
	var ctx context.Context
	if f.FirstArgCtx {
		// TODO: is *HttpRequest a *Context?

		// NOTE: ctx can be nil when doing InspectPC
		argCtx := reflect.ValueOf(args[0]).Elem().Interface()
		if argCtx != nil {
			ctx = argCtx.(context.Context)
		}
		// ctx = *(args[0].(*context.Context))
	} else if f.Closure {
		if len(args) > 0 {
			argCtx, ok := reflect.ValueOf(args[0]).Elem().Interface().(context.Context)
			if ok {
				// modify on the fly
				f.FirstArgCtx = true
				ctx = argCtx
			}
		}
	}
	var perr *error
	if f.LastResultErr {
		perr = results[len(results)-1].(*error)
	} else if f.Closure {
		if len(results) > 0 {
			resErr, ok := reflect.ValueOf(results[0]).Interface().(*error)
			if ok {
				f.LastResultErr = true
				perr = resErr
			}
		}
	}

	// TODO: set FirstArgCtx and LastResultErr
	req := make(object, 0, len(args))
	result := make(object, 0, len(results))
	if f.RecvType != "" {
		req = append(req, field{
			name:   f.RecvName,
			valPtr: recv,
		})
	}
	if !f.FirstArgCtx {
		req = appendFields(req, args, f.ArgNames)
	} else {
		argNames := f.ArgNames
		if argNames != nil {
			argNames = argNames[1:]
		}
		req = appendFields(req, args[1:], argNames)
	}

	var resObject core.Object
	if !f.LastResultErr {
		result = appendFields(result, results, f.ResNames)
		resObject = result
	} else {
		resNames := f.ResNames
		var errName string
		if resNames != nil {
			errName = resNames[len(resNames)-1]
			resNames = resNames[:len(resNames)-1]
		}
		errField := field{
			name:   errName,
			valPtr: results[len(results)-1],
		}
		result = appendFields(result, results[:len(results)-1], resNames)
		resObject = &objectWithErr{
			object: result,
			err:    errField,
		}
	}

	// TODO: what about inlined func?
	// funcArgs := &FuncArgs{
	// 	Recv:    recv,
	// 	Args:    args,
	// 	Results: results,
	// }

	// NOTE: context.TODO() is a constant
	if ctx == nil {
		ctx = context.TODO()
	}

	abortIdx := -1
	n := len(interceptors)
	dataList := make([]interface{}, n)
	for i := n - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		if interceptor.Pre == nil {
			continue
		}
		// if
		data, err := interceptor.Pre(ctx, f, req, resObject)
		dataList[i] = data
		if err != nil {
			if err == ErrAbort {
				abortIdx = i
				// aborted
				break
			}
			// handle error gracefully
			if perr != nil {
				*perr = err
				return nil, true
			} else {
				panic(err)
			}
		}
	}
	if abortIdx >= 0 {
		// run Post immediately
		for i := abortIdx; i < n; i++ {
			interceptor := interceptors[i]
			if interceptor.Post == nil {
				continue
			}
			err := interceptor.Post(ctx, f, req, resObject, dataList[i])
			if err != nil {
				if err == ErrAbort {
					return nil, true
				}
				if perr != nil {
					*perr = err
					return nil, true
				} else {
					panic(err)
				}
			}
		}
		return nil, true
	}

	return func() {
		dispose := setTrappingMark()
		if dispose == nil {
			return
		}
		defer dispose()
		for i := 0; i < n; i++ {
			interceptor := interceptors[i]
			if interceptor.Post == nil {
				continue
			}
			err := interceptor.Post(ctx, f, req, resObject, dataList[i])
			if err != nil {
				if err == ErrAbort {
					return
				}
				if perr != nil {
					*perr = err
					return
				} else {
					panic(err)
				}
			}
		}
	}, false
}

func GetTrappingPC() uintptr {
	key := uintptr(__xgo_link_getcurg())
	val, ok := trappingPC.Load(key)
	if !ok {
		return 0
	}
	return val.(uintptr)
}
func setTrappingMark() func() {
	key := uintptr(__xgo_link_getcurg())
	_, trapping := trappingMark.LoadOrStore(key, struct{}{})
	if trapping {
		return nil
	}
	return func() {
		trappingMark.Delete(key)
	}
}

func clearTrappingMark() {
	key := uintptr(__xgo_link_getcurg())
	trappingMark.Delete(key)
}

func setTrappingPC(pc uintptr) {
	key := uintptr(__xgo_link_getcurg())
	trappingPC.Store(key, pc)
}

func clearTrappingPC() {
	key := uintptr(__xgo_link_getcurg())
	trappingPC.Delete(key)
}
