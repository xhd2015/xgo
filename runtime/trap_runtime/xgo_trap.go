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

func __xgo_set_trap(trap func(pkgPath string, identityName string, generic bool, pc uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	if __xgo_trap_impl != nil {
		panic("trap already set by other packages")
	}
	// ensure this init is called before main
	// we do not care init here, we try our best
	__xgo_trap_impl = trap
}

type __xgo_func_info struct {
	pkgPath      string
	fn           interface{}
	generic      bool
	recvTypeName string
	recvPtr      bool
	name         string
	identityName string // name without pkgPath
	recvName     string
	argNames     []string
	resNames     []string
	firstArgCtx  bool // first argument is context.Context or sub type?
	lastResErr   bool // last res is error or sub type?
}

var funcs []*__xgo_func_info

func __xgo_register_func(pkgPath string, fn interface{}, recvTypeName string, recvPtr bool, name string, identityName string, generic bool, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool) {
	// type intf struct {
	// 	_  uintptr
	// 	pc *uintptr
	// }
	// v := (*intf)(unsafe.Pointer(&fn))
	// fnVal := findfunc(*v.pc)
	// funcName := fnVal.datap.funcName(fnVal.nameOff)
	// println("register func:", funcName)
	funcs = append(funcs, &__xgo_func_info{
		pkgPath:      pkgPath,
		fn:           fn,
		generic:      generic,
		recvTypeName: recvTypeName,
		recvPtr:      recvPtr,
		name:         name,
		identityName: identityName,
		recvName:     recvName,
		argNames:     argNames,
		resNames:     resNames,
		firstArgCtx:  firstArgCtx,
		lastResErr:   lastResErr,
	})
}

func __xgo_for_each_func(f func(pkgPath string, recvTypeName string, recvPtr bool, name string, identityName string, generic bool, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string, firstArgCtx bool, lastResErr bool)) {
	for _, fn := range funcs {
		var pc uintptr

		if !fn.generic {
			type intf struct {
				_  uintptr
				pc *uintptr
			}
			v := (*intf)(unsafe.Pointer(&fn.fn))
			pc = *v.pc
			// fnVal := findfunc(pc)
			// funcName = fnVal.datap.funcName(fnVal.nameOff)
		}
		f(fn.pkgPath, fn.recvTypeName, fn.recvPtr, fn.name, fn.identityName, fn.generic, pc, fn.fn, fn.recvName, fn.argNames, fn.resNames, fn.firstArgCtx, fn.lastResErr)
	}
}

var __xgo_is_init_finished bool

func __xgo_init_finished() bool {
	return __xgo_is_init_finished
}

var __xgo_on_init_finished_callbacks []func()

func __xgo_on_init_finished(fn func()) {
	__xgo_on_init_finished_callbacks = append(__xgo_on_init_finished_callbacks, fn)
}

var __xgo_on_goexits []func()

func __xgo_on_goexit(fn func()) {
	__xgo_on_goexits = append(__xgo_on_goexits, fn)
}

// func GetFuncs_Requires_Xgo() []interface{} {
// 	return funcs
// }
// func GetMethods_Requires_Xgo() []interface{} {
// 	return methods
// }

// func Getcallerpc() uintptr {
// 	return getcallerpc()
// }
// func GetcallerFuncPC() uintptr {
// 	pc := getcallerpc()
// 	fn := findfunc(pc)
// 	return fn.entry()
// }
