package mock

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

// linked by compiler
func __xgo_link_mem_equal(a, b unsafe.Pointer, size uintptr) bool {
	return false
}

var ErrCallOld = errors.New("mock: call old")

type Interceptor func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error

var interceptors []Interceptor

func GetInterceptors() []Interceptor {
	return interceptors
}

// Mock setup mock on given function `fn`.
// `fn` can be a function or a method,
// when `fn` is a method, only the bound
// instance will be mocked, other instances
// are not affected.
// The returned function can be used to cancel
// the passed interceptor.
func Mock(fn interface{}, interceptor Interceptor) func() {
	return mock(fn, interceptor)
}

// Deprecated: use Mock instead
func AddFuncInterceptor(fn interface{}, interceptor Interceptor) func() {
	return mock(fn, interceptor)
}

// TODO: ensure them run in last?
// no abort, run mocks
// mocks are special in that they on run in pre stage
func mock(fn interface{}, interceptor Interceptor) func() {
	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		panic(fmt.Errorf("mock requires func, actual: %v", rv.Kind().String()))
	}
	pc := rv.Pointer()
	fnName := runtime.FuncForPC(pc).Name()

	const fmSuffix = "-fm"
	isMethod := strings.HasSuffix(fnName, fmSuffix)
	var fnNamePrefix string
	if isMethod {
		fnNamePrefix = fnName[:len(fnName)-len(fmSuffix)]
	}

	return trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.PC == 0 {
				// no match, continue
				return nil, nil
			}
			var match bool
			if isMethod {
				// runtime.
				var recvField core.Field
				if runtime.FuncForPC(f.PC).Name() == fnNamePrefix && f.RecvType != "" {
					// the first field is recv
					recvField = args.GetFieldIndex(0)
				}
				var recvPtr interface{}
				if recvField != nil {
					recvPtr = recvField.Ptr()
				}
				if recvPtr != nil {
					match = isSameBoundMethod(recvPtr, fn)
				}
			} else if f.PC == pc {
				match = true
			}
			if !match {
				// continue
				return nil, nil
			}

			// TODO: add panic check
			err = interceptor(ctx, f, args, result)
			if err == ErrCallOld {
				// continue
				return nil, nil
			}
			// when match func, default to use mock
			return nil, trap.ErrAbort
		},
	})
}

func CallOld() {
	// TODO: implement recover
	panic(ErrCallOld)
}

// this function checks if the given
// `recvPtr` has the same value compared
// to the given `methodValue`.
// The `methodValue` should be passed as
// `file.Write“.
func isSameBoundMethod(recvPtr interface{}, methodValue interface{}) bool {
	// can also be a constant
	// size := unsafe.Sizeof(*(*large)(nil))
	size := reflect.TypeOf(recvPtr).Elem().Size()
	type _intfRecv struct {
		_    uintptr // type word
		data *byte   // data word
	}

	a := (*_intfRecv)(unsafe.Pointer(&recvPtr))
	type _methodValue struct {
		_    uintptr // pc
		recv byte
	}
	type _intf struct {
		_    uintptr // type word
		data *_methodValue
	}
	ppb := (*_intf)(unsafe.Pointer(&methodValue))
	pb := *ppb
	b := unsafe.Pointer(&pb.data.recv)

	return __xgo_link_mem_equal(unsafe.Pointer(a.data), b, size)
}

// mock context
// MockContext
// MockPoint
// MockController
// MockRecorder
// type MockContext interface {
// 	Log()
// }

// a mock context seems not needed, because all can be done by calling
// into mock's function
