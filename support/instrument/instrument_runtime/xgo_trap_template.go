// This file should be inserted into runtime to provide hack points
// the 'go:build ignore' line will be removed when copy to goroot

//go:build ignore

package runtime

import (
	"unsafe"
)

// make unsafe always imported
var _ = unsafe.Pointer(nil)

// this signature is to avoid relying on defined type
// just raw string,interface{},[]string, []interface{} are needed
var __xgo_trap func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool)

var __xgo_var_trap func(name string, varAddr interface{}, res interface{})
var __xgo_var_ptr_trap func(name string, varAddr interface{}, res interface{})

func XgoTrap(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool) {
	if __xgo_trap == nil {
		return nil, false
	}
	return __xgo_trap(recvName, recvPtr, argNames, args, resultNames, results)
}

func XgoTrapVar(name string, varAddr interface{}, res interface{}) {
	if __xgo_var_trap == nil {
		return
	}
	__xgo_var_trap(name, varAddr, res)
}

func XgoTrapVarPtr(name string, varAddr interface{}, res interface{}) {
	if __xgo_var_ptr_trap == nil {
		return
	}
	__xgo_var_ptr_trap(name, varAddr, res)
}

func XgoSetTrap(trap func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool)) {
	if __xgo_trap != nil {
		panic("__xgo_trap already set")
	}
	__xgo_trap = trap
}

func XgoSetVarTrap(trap func(name string, varAddr interface{}, res interface{})) {
	if __xgo_var_trap != nil {
		panic("__xgo_var_trap already set")
	}
	__xgo_var_trap = trap
}

func XgoSetVarPtrTrap(trap func(name string, varAddr interface{}, res interface{})) {
	if __xgo_var_ptr_trap != nil {
		panic("__xgo_var_ptr_trap already set")
	}
	__xgo_var_ptr_trap = trap
}

// __xgo_g is a wrapper around the runtime.G struct
// to avoid exposing the runtime.G struct to the user
// and to avoid having to import the runtime package
// in the user's code.
type __xgo_g struct {
	gls                 map[interface{}]interface{}
	looseJsonMarshaling bool
}

func XgoGetCurG() unsafe.Pointer {
	curg := getg().m.curg
	// println("get g:", hex(uintptr(unsafe.Pointer(curg))))
	return unsafe.Pointer(&curg.__xgo_g)
}

// Peek panic without recover
// check gorecover() for implementation details
func XgoPeekPanic() interface{} {
	gp := getg()
	p := gp._panic
	if p == nil || p.goexit || p.recovered {
		return nil
	}
	return p.arg
}

func XgoIsLooseJsonMarshaling() bool {
	curg := getg().m.curg
	return curg.__xgo_g.looseJsonMarshaling
}

// XgoGetFullPCName returns full name
// without ellipsis
func XgoGetFullPCName(pc uintptr) string {
	return __xgo_get_pc_name_impl(pc)
}

// goroutine creation
var __xgo_on_create_g_callbacks []func(g unsafe.Pointer, childG unsafe.Pointer)
var __xgo_on_exit_g_callbacks []func()

func XgoOnCreateG(callback func(g unsafe.Pointer, childG unsafe.Pointer)) {
	__xgo_on_create_g_callbacks = append(__xgo_on_create_g_callbacks, callback)
}

func XgoOnExitG(callback func()) {
	__xgo_on_exit_g_callbacks = append(__xgo_on_exit_g_callbacks, callback)
}

func __xgo_callback_on_create_g(curg *g, newg *g) {
	if newg == nil {
		return
	}
	// newg might be reused from an already exited goroutine
	// so here we need to explicity clear the __xgo_g
	// clear
	newg.__xgo_g = __xgo_g{}
	if len(__xgo_on_create_g_callbacks) == 0 {
		return
	}
	if curg == nil {
		return
	}
	// println("new g:", hex(uintptr(unsafe.Pointer(newg))), "from g:", hex(uintptr(unsafe.Pointer(curg))))
	curg_p := unsafe.Pointer(&curg.__xgo_g)
	newg_p := unsafe.Pointer(&newg.__xgo_g)
	for _, callback := range __xgo_on_create_g_callbacks {
		callback(curg_p, newg_p)
	}
}

func __xgo_callback_on_exit_g() {
	for _, callback := range __xgo_on_exit_g_callbacks {
		callback()
	}
}
