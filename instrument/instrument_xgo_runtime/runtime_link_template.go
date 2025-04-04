// This file should be inserted into xgo's runtime to provide hack points
// the 'go:build ignore' line will be removed when replaced by overlay

//go:build ignore

package runtime

import (
	"runtime"
	"unsafe"
)

func XgoSetTrap(trap func(info unsafe.Pointer, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	runtime.XgoSetTrap(trap)
}

func XgoSetVarTrap(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
	runtime.XgoSetVarTrap(trap)
}

func XgoSetVarPtrTrap(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
	runtime.XgoSetVarPtrTrap(trap)
}

func XgoGetCurG() unsafe.Pointer {
	//
	return runtime.XgoGetCurG()
}

func XgoPeekPanic() interface{} {
	//
	return runtime.XgoPeekPanic()
}

func XgoGetFullPCName(pc uintptr) string {
	//
	return runtime.XgoGetFullPCName(pc)
}

func XgoOnCreateG(callback func(g unsafe.Pointer, childG unsafe.Pointer)) {
	runtime.XgoOnCreateG(callback)
}

func XgoOnExitG(callback func()) {
	runtime.XgoOnExitG(callback)
}
