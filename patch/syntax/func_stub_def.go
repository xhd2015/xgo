package syntax

const expected__xgo_stub_def = `struct {
	PkgPath      string
	Fn           interface{}
	PC           uintptr // filled later
	Interface    bool
	Generic      bool
	Closure      bool // is the given function a closure
	RecvTypeName string
	RecvPtr      bool
	Name         string
	IdentityName string // name without pkgPath

	RecvName string
	ArgNames []string
	ResNames []string

	// can be retrieved at runtime
	FirstArgCtx bool // first argument is context.Context or sub type?
	LastResErr  bool // last res is error or sub type?

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
// TODO: if we encounter any of these name conflics, we skip this file or use a different name
const helperCode = helperCodeGen
