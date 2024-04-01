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
		__xgo_link_set_trap(trapImpl)
	})
}

func __xgo_link_set_trap(trapImpl func(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_set_trap.(xgo required)")
}

// Skip serves as mark to tell xgo not insert
// trap instructions for the function that
// calls Skip()
// NOTE: the function body is intenionally leave empty
// as trap.Skip() is just a mark that makes
// sense at compile time.
func Skip() {}

func GetTrappingPC() uintptr {
	key := uintptr(__xgo_link_getcurg())
	val, ok := trappingPC.Load(key)
	if !ok {
		return 0
	}
	return val.(uintptr)
}

var trappingMark sync.Map // <goroutine key> -> struct{}{}
var trappingPC sync.Map   // <gorotuine key> -> PC

// link to runtime
// xgo:notrap
func trapImpl(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	dispose := setTrappingMark()
	if dispose == nil {
		return nil, false
	}
	defer dispose()

	// setup context
	setTrappingPC(pc)
	defer clearTrappingPC()

	type intf struct {
		_  uintptr
		pc *uintptr
	}
	interceptors := GetAllInterceptors()
	n := len(interceptors)
	if n == 0 {
		return nil, false
	}
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
		//
		return nil, false
	}
	if f.RecvType != "" && methodHasBeenTrapped && recv == nil {
		// let go to the next interceptor
		return nil, false
	}

	// retrieve context
	var ctx context.Context
	if f.FirstArgCtx {
		// TODO: is *HttpRequest a *Context?
		ctx = reflect.ValueOf(args[0]).Elem().Interface().(context.Context)
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
			resErr, ok := reflect.ValueOf(args[0]).Interface().(*error)
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
