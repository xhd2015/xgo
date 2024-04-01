// This file should be inserted into runtime to provide hack points
// the 'go:build ignore' line will be removed when copy to goroot

//go:build ignore

package runtime

import (
	"unsafe"
)

// use getg().m.curg instead of getg()
// see: https://github.com/golang/go/blob/master/src/runtime/HACKING.md
func __xgo_getcurg() unsafe.Pointer { return unsafe.Pointer(getg().m.curg) }

// exported so other func can call it
var __xgo_trap_impl func(pkgPath string, identityName string, generic bool, funcPC uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)

// this is so elegant that you cannot ignore it
func __xgo_trap(pkgPath string, identityName string, generic bool, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	if __xgo_trap_impl == nil {
		return nil, false
	}
	pc := getcallerpc()
	fn := findfunc(pc)
	// TODO: what about inlined func?
	// funcName := fn.datap.funcName(fn.nameOff) // not necessary,because it is unsafe
	return __xgo_trap_impl(pkgPath, identityName, generic, fn.entry() /*>=go1.18*/, recv, args, results)
}

func __xgo_trap_for_generated(pkgPath string, pc uintptr, identityName string, generic bool, recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	if __xgo_trap_impl == nil {
		return nil, false
	}
	fn := findfunc(pc)
	// TODO: what about inlined func?
	// funcName := fn.datap.funcName(fn.nameOff) // not necessary,because it is unsafe
	return __xgo_trap_impl(pkgPath, identityName, generic, fn.entry() /*>=go1.18*/, recv, args, results)
}

func __xgo_set_trap(trap func(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	if __xgo_trap_impl != nil {
		panic("trap already set by other packages")
	}
	// ensure this init is called before main
	// we do not care init here, we try our best
	__xgo_trap_impl = trap
}

// NOTE: runtime has problem when using slice
var __xgo_registered_func_infos []interface{}

// a function cannot have too many params, so use a struct to wrap them
// the client should use reflect to retrieve these fields respectively

func __xgo_register_func(info interface{}) {
	__xgo_registered_func_infos = append(__xgo_registered_func_infos, info)
}

func __xgo_retrieve_all_funcs_and_clear(f func(info interface{})) {
	for _, fn := range __xgo_registered_func_infos {
		f(fn)
	}
	// clear
	__xgo_registered_func_infos = nil
}

var __xgo_is_init_finished bool

func __xgo_init_finished() bool {
	return __xgo_is_init_finished
}

var __xgo_on_init_finished_callbacks []func()

func __xgo_on_init_finished(fn func()) {
	__xgo_on_init_finished_callbacks = append(__xgo_on_init_finished_callbacks, fn)
}

// goroutine creates and exits callbacks
var __xgo_on_gonewproc_callbacks []func(g uintptr)
var __xgo_on_goexits []func()

func __xgo_on_gonewproc(fn func(g uintptr)) {
	__xgo_on_gonewproc_callbacks = append(__xgo_on_gonewproc_callbacks, fn)
}

func __xgo_on_goexit(fn func()) {
	__xgo_on_goexits = append(__xgo_on_goexits, fn)
}

var __xgo_on_test_starts []interface{} // func(t *testing.T,fn func(t *testing.T))

func __xgo_on_test_start(fn interface{}) {
	__xgo_on_test_starts = append(__xgo_on_test_starts, fn)
}

func __xgo_get_test_starts() []interface{} {
	return __xgo_on_test_starts
}

// check gorecover() for implementation details
func __xgo_peek_panic() interface{} {
	gp := getg()
	p := gp._panic
	if p == nil || p.goexit || p.recovered {
		return nil
	}
	return p.arg
}

func __xgo_mem_equal(a, b unsafe.Pointer, size uintptr) bool {
	return memequal(a, b, size)
}

func __xgo_get_pc_name(pc uintptr) string {
	return __xgo_get_pc_name_impl(pc)
}

// set by reflect: not enabled
// var __xgo_get_all_method_by_name_impl func(v interface{}, name string) interface{}

// func __xgo_set_all_method_by_name_impl(fn func(v interface{}, name string) interface{}) {
// 	__xgo_get_all_method_by_name_impl = fn
// }

// func __xgo_get_all_method_by_name(v interface{}, name string) interface{} {
// 	return __xgo_get_all_method_by_name_impl(v, name)
// }

// func Getcallerpc() uintptr {
// 	return getcallerpc()
// }
// func GetcallerFuncPC() uintptr {
// 	pc := getcallerpc()
// 	fn := findfunc(pc)
// 	return fn.entry()
// }
