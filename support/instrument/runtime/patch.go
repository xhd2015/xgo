package runtime

const RuntimeGetFuncName_Go117_120 = `
func __xgo_get_pc_name_impl(pc uintptr) string {
	return FuncForPC(pc).Name()
}
`

// start with go1.21, the runtime.FuncForPC(pc).Name()
// is wrapped in funcNameForPrint(...), we unwrap it.
// it is confirmed that in go1.21,go1.22 and go1.23,
// the name is wrapped.
// NOTE: when upgrading to go1.24, should check
// the implementation again
const RuntimeGetFuncName_Go121 = `
func __xgo_get_pc_name_impl(pc uintptr) string {
	return FuncForPC(pc).__xgo_no_print_name()
}

func (f *Func) __xgo_no_print_name() string {
	if f == nil {
		return ""
	}
	fn := f.raw()
	if fn.isInlined() { // inlined version
		fi := (*funcinl)(unsafe.Pointer(fn))
		return fi.name
	}
	return funcname(f.funcInfo())
}
`

// FuncForPC(pc).Name() just works fine, there is no
// string split
const RuntimeGetFuncName_Go120_Unused = `
func __xgo_get_pc_name_impl(pc uintptr) string {
	fn := findfunc(pc)
	return fn.datap.funcName(fn.nameOff)
}
// workaround for go1.20, go1.21 will including this by go
func (md *moduledata) funcName(nameOff int32) string {
	if nameOff == 0 {
		return ""
	}
	return gostringnocopy(&md.funcnametab[nameOff])
}
`

const RuntimeProcGoroutineCreatedPatch = `for _, fn := range __xgo_on_gonewproc_callbacks {
	fn(uintptr(unsafe.Pointer(xgo_newg)))
}
`
