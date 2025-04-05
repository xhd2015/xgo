package info

// NOTE: this info package should remain extremely lightweight,
// so that it can be used by stdlib
// DO NOT IMPORT ANYTHING except runtime

// this package is not placed in internal
// because it will be used by user code and stdlib

type Kind int

const (
	Kind_Func   Kind = 0
	Kind_Var    Kind = 1
	Kind_VarPtr Kind = 2
	Kind_Const  Kind = 3
)

var stagingFuncs []*Func
var registerHandler func(fn *Func)

func SetupRegisterHandler(register func(fn *Func)) {
	if registerHandler != nil {
		panic("registerHandler already set")
	}
	registerHandler = register
	saved := stagingFuncs
	stagingFuncs = nil
	for _, fn := range saved {
		register(fn)
	}
}

func Register(fn *Func) {
	if registerHandler != nil {
		registerHandler(fn)
		return
	}
	stagingFuncs = append(stagingFuncs, fn)
}

type Func struct {
	Kind Kind
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
	// last last result error
	LastResultErr bool
}

func (c *Func) DisplayName() string {
	if c.RecvType != "" {
		return c.RecvType + "." + c.Name
	}
	return c.Name
}

func (c Kind) String() string {
	switch c {
	case Kind_Func:
		return "func"
	case Kind_Var:
		return "var"
	case Kind_VarPtr:
		return "var_ptr"
	case Kind_Const:
		return "const"
	default:
		print("unrecognized kind: ")
		panic(int(c))
	}
}
