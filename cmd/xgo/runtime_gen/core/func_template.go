//go:build ignore

package core

// ==start xgo func==
type __PREFIX__Kind int

const (
	__PREFIX__Kind_Func   __PREFIX__Kind = 0
	__PREFIX__Kind_Var    __PREFIX__Kind = 1
	__PREFIX__Kind_VarPtr __PREFIX__Kind = 2
	__PREFIX__Kind_Const  __PREFIX__Kind = 3
)

type __PREFIX__FuncInfo struct {
	Kind __PREFIX__Kind
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
