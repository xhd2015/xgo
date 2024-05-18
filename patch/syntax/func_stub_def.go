package syntax

const expected__xgo_stub_def = `struct {
	PkgPath      string
	Kind         int // 0 = func, 1 = var, 2=var_ptr 3 = const
	Fn           interface{}
	Var          interface{} // pointer to a variable if this is a declare variable
	PC           uintptr     // filled later
	Interface    bool
	Generic      bool
	Closure      bool // is the given function a closure
	Stdlib       bool
	RecvTypeName string
	RecvPtr      bool
	Name         string
	IdentityName string // name without pkgPath

	RecvName string
	ArgNames []string
	ResNames []string

	// Deprecated
	// these two fields can be retrieved at runtime
	FirstArgCtx bool // first argument is context.Context or sub type?
	// Deprecated
	LastResErr bool // last res is error or sub type?

	File string
	Line int
}`

func init() {
	if expected__xgo_stub_def != __xgo_stub_def {
		panic("__xgo_stub_def signature changed, run go generate and update expected__xgo_stub_def correspondly")
	}
}

// helperCode: helps us(xgo) write IR more easily
// NOTE: normal code should never define names starts with __xgo
// TODO: if we encounter any of these name conflicts, we skip this file or use a different name
const helperCode = helperCodeGen
