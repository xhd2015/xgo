package mock

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/legacy"
	"github.com/xhd2015/xgo/runtime/trap"
	trap_legacy "github.com/xhd2015/xgo/runtime/trap/legacy"
)

type Interceptor = trap.Interceptor

// Deprecated: use Mock instead
func MockByName(pkgPath string, funcName string, interceptor Interceptor) func() {
	if !legacy.V1_0_0 {
		panic("MockByName is deprecated and no longer supported, use Mock instead")
	}
	recv, fn, funcPC, trappingPC := getFuncByName(pkgPath, funcName)
	return mockLegacy(recv, fn, funcPC, trappingPC, interceptor)
}

// Can instance be nil?
// Deprecated: use Mock instead
func MockMethodByName(instance interface{}, method string, interceptor Interceptor) func() {
	if !legacy.V1_0_0 {
		panic("MockMethodByName is deprecated and no longer supported, use Mock instead")
	}
	recvPtr, fn, funcPC, trappingPC := getMethodByName(instance, method)
	return mockLegacy(recvPtr, fn, funcPC, trappingPC, interceptor)
}

func getFunc(fn interface{}) (recvPtr interface{}, fnInfo *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	// if the target function is a method, then a
	// recv ptr must be given
	recvPtr, fnInfo, funcPC, trappingPC = trap_legacy.InspectPC(fn)
	if fnInfo == nil {
		pc := reflect.ValueOf(fn).Pointer()
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			panic(fmt.Errorf("failed to setup mock for variable: 0x%x", pc))
		}
		panic(fmt.Errorf("failed to setup mock for: %v", fn.Name()))
	}
	return recvPtr, fnInfo, funcPC, trappingPC
}

func getFuncByName(pkgPath string, funcName string) (recvPtr interface{}, fn *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	fn = functab.GetFuncByPkg(pkgPath, funcName)
	if fn == nil {
		panic(fmt.Errorf("failed to setup mock for: %s.%s", pkgPath, funcName))
	}
	return nil, fn, 0, 0
}

func getMethodByName(instance interface{}, method string) (recvPtr interface{}, fn *core.FuncInfo, funcPC uintptr, trappingPC uintptr) {
	// extract instance's reflect.Type
	// use that type to query for reflect mapping in functab:
	//    reflectTypeMapping map[reflect.Type]map[string]*funcInfo
	t := reflect.TypeOf(instance)
	funcMapping := functab.GetTypeMethods(t)
	if funcMapping == nil {
		panic(fmt.Errorf("failed to setup mock for type %T", instance))
	}
	fn = funcMapping[method]
	if fn == nil {
		panic(fmt.Errorf("failed to setup mock for: %T.%s", instance, method))
	}

	addr := reflect.New(t)
	addr.Elem().Set(reflect.ValueOf(instance))

	return addr.Interface(), fn, 0, 0
}

// TODO: ensure them run in last?
// no abort, run mocks
// mocks are special in that they on run in pre stage

// if mockFnInfo is a function, mockRecvPtr is always nil
// if mockFnInfo is a method,
//   - if mockRecvPtr has a value, then only call to that instance will be mocked
//   - if mockRecvPtr is nil, then all call to the function will be mocked
//
// Deprecated: use mock instead
func mockLegacy(mockRecvPtr interface{}, mockFnInfo *core.FuncInfo, funcPC uintptr, trappingPC uintptr, interceptor Interceptor) func() {
	panic("Deprecated: should be not called")
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
