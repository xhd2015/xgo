// This file should be inserted into runtime to provide hack points

//go:build ignore

package runtime

import (
	"unsafe"
)

// important thing here:
//
//		get arg types and argNames
//	 get names
//
// two ways to
//
//	by function name
//
// it needs defer
//
//	if something{
//	    next()
//	    myAction()
//	}

// use getg().m.curg instead of getg()
// see: https://github.com/golang/go/blob/master/src/runtime/HACKING.md
func __xgo_getcurg() unsafe.Pointer { return unsafe.Pointer(getg().m.curg) }

// exported so other func can call it
var TrapImpl_Requires_Xgo func(funcName string, funcPC uintptr, recv interface{}, args []interface{}, results []interface{}) (func(), bool)

// this is so elegant that you cannot ignore it
func __xgo_trap(recv interface{}, args []interface{}, results []interface{}) (func(), bool) {
	if TrapImpl_Requires_Xgo == nil {
		return nil, false
	}
	pc := getcallerpc()
	fn := findfunc(pc)
	// TODO: what about inlined func?
	funcName := fn.datap.funcName(fn.nameOff)
	return TrapImpl_Requires_Xgo(funcName, fn.entry(), recv, args, results)
}

type __xgo_func_info struct {
	pkgPath  string
	fn       interface{}
	recvName string
	argNames []string
	resNames []string
}

var funcs []*__xgo_func_info

func __xgo_register_func(pkgPath string, fn interface{}, recvName string, argNames []string, resNames []string) {
	// type intf struct {
	// 	_  uintptr
	// 	pc *uintptr
	// }
	// v := (*intf)(unsafe.Pointer(&fn))
	// fnVal := findfunc(*v.pc)
	// funcName := fnVal.datap.funcName(fnVal.nameOff)
	// println("register func:", funcName)
	funcs = append(funcs, &__xgo_func_info{
		pkgPath:  pkgPath,
		fn:       fn,
		recvName: recvName,
		argNames: argNames,
		resNames: resNames,
	})
}

// func RegisterFunc_Requires_Xgo(fn interface{}, recvName string, argNames []string, resNames []string) {
// 	__xgo_register_func(fn, recvName, argNames, resNames)
// }

func __xgo_for_each_func(f func(pkgPath string, funcName string, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string)) {
	for _, fn := range funcs {
		type intf struct {
			_  uintptr
			pc *uintptr
		}
		v := (*intf)(unsafe.Pointer(&fn.fn))
		pc := *v.pc
		fnVal := findfunc(pc)
		funcName := fnVal.datap.funcName(fnVal.nameOff)
		f(fn.pkgPath, funcName, pc, fn.fn, fn.recvName, fn.argNames, fn.resNames)
	}
}

// func GetFuncs_Requires_Xgo() []interface{} {
// 	return funcs
// }
// func GetMethods_Requires_Xgo() []interface{} {
// 	return methods
// }

func Getcallerpc() uintptr {
	return getcallerpc()
}
func GetcallerFuncPC() uintptr {
	pc := getcallerpc()
	fn := findfunc(pc)
	return fn.entry()
}
