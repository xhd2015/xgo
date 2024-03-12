package trap

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"unsafe"
)

var used bool

func Use() {
	if used {
		return
	}
	if runtime.TrapImpl_Requires_Xgo != nil {
		panic(errors.New("trap already set by other packages"))
	}
	// ensure this init is called before main
	// we do not care init here, we try our best
	runtime.TrapImpl_Requires_Xgo = trapImpl
	used = true
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
	// check if the calling func is an interceptor, if so, skip
	n := len(interceptors)
	if n == 0 {
		return nil, false
	}
	if false {
		// don't do manual check
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

	pkgPath, recvName, recvPtr, funcShortName := parseFuncName(funcName)
	f := &FuncInfo{
		Pkg:      pkgPath,
		RecvName: recvName,
		RecvPtr:  recvPtr,
		Name:     funcShortName,
		FullName: funcName,
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

// a/b/c.A
// a/b/c.(*C).X
// a/b/c.C.Y
// a/b/c.Z
func parseFuncName(fullName string) (pkgPath string, recvName string, recvPtr bool, funcName string) {
	s := fullName
	funcNameDot := strings.LastIndex(s, ".")
	if funcNameDot < 0 {
		funcName = s
		return
	}
	funcName = s[funcNameDot+1:]
	s = s[:funcNameDot]

	recvDot := strings.LastIndex(s, ".")
	if recvDot < 0 {
		pkgPath = s
		return
	}
	recvName = s[recvDot+1:]
	s = s[:recvDot]
	recvName = strings.TrimPrefix(recvName, "(")
	recvName = strings.TrimSuffix(recvName, ")")
	if strings.HasPrefix(recvName, "*") {
		recvPtr = true
		recvName = recvName[1:]
	}

	pkgPath = s

	return
}
