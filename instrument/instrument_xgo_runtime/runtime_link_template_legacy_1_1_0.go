// This file should be inserted into xgo's runtime to provide hack points
// the 'go:build ignore' line will be removed when replaced by overlay

//go:build ignore

package runtime

import (
	// "runtime"
	"time"
	"unsafe"
)

func init() {
	__xgo_init_legacy_v1_1_0()
}

func XgoSetTrap(trap func(info unsafe.Pointer, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool)) {
	runtime_link_XgoSetTrap(trap)
}

func XgoSetVarTrap(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
	runtime_link_XgoSetVarTrap(trap)
}

func XgoSetVarPtrTrap(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
	runtime_link_XgoSetVarPtrTrap(trap)
}
func XgoSetupRegisterHandler(register func(fn unsafe.Pointer)) {
	runtime_link_XgoSetupRegisterHandler(register)
}

func XgoGetCurG() unsafe.Pointer {
	return runtime_link_XgoGetCurG()
}

func XgoPeekPanic() (interface{}, uintptr) {
	return runtime_link_XgoPeekPanic()
}

func XgoGetFullPCName(pc uintptr) string {
	return runtime_link_XgoGetFullPCName(pc)
}

func XgoOnCreateG(callback func(g unsafe.Pointer, childG unsafe.Pointer)) {
	runtime_link_XgoOnCreateG(callback)
}

func XgoOnExitG(callback func()) {
	runtime_link_XgoOnExitG(callback)
}

// XgoRealTimeNow returns the true time.Now()
// this will be rewritten to time.XgoRealNow() if time.Now was rewritten
func XgoRealTimeNow() time.Time {
	return time.XgoRealNow()
}

func XgoOnInitFinished(callback func()) {
	runtime_link_XgoOnInitFinished(callback)
}

func XgoInitFinished() bool {
	return runtime_link_XgoInitFinished()
}

// workarounds

//go:noinline
func __xgo_init_legacy_v1_1_0() {
	runtime_link_XgoSetTrap = runtime_link_XgoSetTrap
	runtime_link_XgoSetVarTrap = runtime_link_XgoSetVarTrap
	runtime_link_XgoSetVarPtrTrap = runtime_link_XgoSetVarPtrTrap
	runtime_link_XgoSetupRegisterHandler = runtime_link_XgoSetupRegisterHandler
	runtime_link_XgoGetCurG = runtime_link_XgoGetCurG
	runtime_link_XgoPeekPanic = runtime_link_XgoPeekPanic
	runtime_link_XgoGetFullPCName = runtime_link_XgoGetFullPCName
	runtime_link_XgoOnCreateG = runtime_link_XgoOnCreateG
	runtime_link_XgoOnExitG = runtime_link_XgoOnExitG
	runtime_link_XgoOnInitFinished = runtime_link_XgoOnInitFinished
	runtime_link_XgoInitFinished = runtime_link_XgoInitFinished
}

var runtime_link_XgoSetTrap = func(trap func(info unsafe.Pointer, recvPtr interface{}, args []interface{}, results []interface{}) (func(), bool)) {
}

var runtime_link_XgoSetVarTrap = func(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
}

var runtime_link_XgoSetVarPtrTrap = func(trap func(info unsafe.Pointer, varAddr interface{}, res interface{})) {
}

var runtime_link_XgoSetupRegisterHandler = func(register func(fn unsafe.Pointer)) {
}

var runtime_link_XgoGetCurG = func() unsafe.Pointer {
	return nil
}

var runtime_link_XgoPeekPanic = func() (interface{}, uintptr) {
	return nil, 0
}

var runtime_link_XgoGetFullPCName = func(pc uintptr) string {
	return ""
}

var runtime_link_XgoOnCreateG = func(callback func(g unsafe.Pointer, childG unsafe.Pointer)) {
}

var runtime_link_XgoOnExitG = func(callback func()) {
}

var runtime_link_XgoOnInitFinished = func(callback func()) {
}

var runtime_link_XgoInitFinished = func() bool {
	return false
}

// append here: `type XgoFuncInfo struct {`
// ==start xgo func==

type XgoKind int

const (
	XgoKind_Func   XgoKind = 0
	XgoKind_Var    XgoKind = 1
	XgoKind_VarPtr XgoKind = 2
	XgoKind_Const  XgoKind = 3
)

type XgoFuncInfo struct {
	Kind XgoKind
	// full name, format: {pkgPath}.{receiver}.{funcName}
	// example:  github.com/xhd2015/xgo/runtime/core.(*SomeType).SomeFunc
	FullName string
	Pkg      string
	// identity name within a package, for ptr-method, it's something like `(*SomeType).SomeFunc`
	// run `go run ./test/example/method` to verify
	IdentityName string
	Name         string
	RecvType     string
	RecvPtr      bool

	// is this an interface method?
	Interface bool

	// is this a generic function?
	Generic bool

	// is this a closure?
	Closure bool

	// is this function from stdlib
	Stdlib bool

	// source info
	File string
	Line int

	// PC is the function entry point
	// if it's a method, it is the underlying entry point
	// for all instances, it is the same
	PC   uintptr     `json:"-"`
	Func interface{} `json:"-"`
	Var  interface{} `json:"-"` // var address

	RecvName string
	ArgNames []string
	ResNames []string

	// is first argument ctx
	FirstArgCtx bool
	// last result error
	LastResultErr bool
}

// ==end xgo func==
