// This file will be replaced by runtime_link_template.go
// These lines are to keep in sync with the template. Don't change the filename.

package runtime

import (
	// "runtime"
	"unsafe"
)

func XgoSetTrap(trap func(info unsafe.Pointer, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	logError("WARNING: failed to link runtime.XgoSetTrap(requires xgo).")
}

func XgoSetVarTrap(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
	logError("WARNING: failed to link runtime.XgoSetVarTrap(requires xgo).")
}

func XgoSetVarPtrTrap(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
	logError("WARNING: failed to link runtime.XgoSetVarPtrTrap(requires xgo).")
}

func XgoGetCurG() unsafe.Pointer {
	logError("WARNING: failed to link runtime.XgoGetCurG(requires xgo).")
	return nil
}

func XgoPeekPanic() interface{} {
	logError("WARNING: failed to link runtime.XgoPeekPanic(requires xgo).")
	return nil
}

func XgoGetFullPCName(pc uintptr) string {
	logError("WARNING: failed to link runtime.XgoGetFullPCName(requires xgo).")
	return ""
}

func XgoOnCreateG(callback func(g unsafe.Pointer, childG unsafe.Pointer)) {
	logError("WARNING: failed to link runtime.XgoOnCreateG(requires xgo).")
}

func XgoOnExitG(callback func()) {
	logError("WARNING: failed to link runtime.XgoOnExitG(requires xgo).")
}
