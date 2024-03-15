package trap

import (
	"context"
	"runtime"
	"sync"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core/functab"
)

var setupOnce sync.Once

func ensureSetupTrap() {
	setupOnce.Do(func() {
		__xgo_link_set_trap(trapImpl)
	})
}

func __xgo_link_set_trap(trapImpl func(funcName string, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	panic("failed to link __xgo_link_set_trap")
}

func Skip() {
	// this is intenionally leave empty
	// as trap.Skip() is a function used
	// to mark the caller should not be trapped.
	// one can also use trap.Skip() in
	// the non-interceptor context
}

// link to runtime
// xgo:notrap
func trapImpl(funcName string, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	type intf struct {
		_  uintptr
		pc *uintptr
	}
	interceptors := GetAllInterceptors()
	n := len(interceptors)
	if n == 0 {
		return nil, false
	}
	if false {
		// check if the calling func is an interceptor, if so, skip
		// UPDATE: don't do manual check
		for i := 0; i < n; i++ {
			if interceptors[i].Pre == nil {
				continue
			}
			ipc := (**uintptr)(unsafe.Pointer(&interceptors[i].Pre))
			pcName := runtime.FuncForPC(**ipc).Name()
			_ = pcName
			if **ipc == pc {
				return nil, false
			}
		}
	}
	f := functab.InfoPC(pc)
	if f == nil {
		// fallback to default
		pkgPath, recvName, recvPtr, funcShortName := functab.ParseFuncName(funcName, true)
		f = &functab.FuncInfo{
			Pkg:      pkgPath,
			RecvName: recvName,
			RecvPtr:  recvPtr,
			Name:     funcShortName,
			FullName: funcName,
		}
	}
	// TODO: what about inlined func?
	funcArgs := &FuncArgs{
		Recv:    recv,
		Args:    args,
		Results: results,
	}

	var perr *error
	if len(results) > 0 {
		if errPtr, ok := results[len(results)-1].(*error); ok {
			perr = errPtr
		}
	}

	// NOTE: ctx may
	var ctx context.Context
	if len(args) > 0 {
		// TODO: is *HttpRequest a *Context?
		if argCtxPtr, ok := args[0].(*context.Context); ok {
			ctx = *argCtxPtr
		}
	}
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
		data, err := interceptor.Pre(ctx, f, funcArgs)
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
		for i := abortIdx; i < n; i++ {
			interceptor := interceptors[i]
			if interceptor.Post == nil {
				continue
			}
			err := interceptor.Post(ctx, f, funcArgs, dataList[i])
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
		for i := 0; i < n; i++ {
			interceptor := interceptors[i]
			if interceptor.Post == nil {
				continue
			}
			err := interceptor.Post(ctx, f, funcArgs, dataList[i])
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
