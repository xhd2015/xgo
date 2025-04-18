// This file should be inserted into xgo's runtime to provide hack points
// the 'go:build ignore' line will be removed when replaced by overlay

//go:build ignore

package runtime

import (
	"runtime"
	"time"
	"unsafe"
)

// this ensures runtime to be always imported
var _ = runtime.NumCPU
var _ = unsafe.Pointer(nil)

type XgoFuncInfo = runtime.XgoFuncInfo

func XgoSetTrap(trap func(info unsafe.Pointer, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	runtime.XgoSetTrap(trap)
}

func XgoSetVarTrap(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
	runtime.XgoSetVarTrap(trap)
}

func XgoSetVarPtrTrap(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
	runtime.XgoSetVarPtrTrap(trap)
}

func XgoSetupRegisterHandler(register func(fn unsafe.Pointer)) {
	runtime.XgoSetupRegisterHandler(register)
}

func XgoGetCurG() unsafe.Pointer {
	//
	return runtime.XgoGetCurG()
}

func XgoPeekPanic() (interface{}, uintptr) {
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

// XgoRealTimeNow returns the true time.Now()
// this will be rewritten to time.XgoRealNow() if time.Now was rewritten
func XgoRealTimeNow() time.Time {
	return time.XgoRealNow()
}

func XgoOnInitFinished(callback func()) {
	runtime.XgoOnInitFinished(callback)
}

func XgoInitFinished() bool {
	//
	return runtime.XgoInitFinished()
}
