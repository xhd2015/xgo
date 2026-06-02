package types

type Kind int

const (
	Kind_Func   Kind = 0
	Kind_Var    Kind = 1
	Kind_VarPtr Kind = 2
	Kind_Const  Kind = 3
)

type FuncInfo struct {
	Kind Kind
	FullName     string
	Pkg          string
	IdentityName string
	Name         string
	RecvType     string
	RecvPtr      bool

	Interface bool
	Generic   bool
	Closure   bool
	Stdlib    bool

	File string
	Line int

	PC   uintptr     `json:"-"`
	Func interface{} `json:"-"`
	Var  interface{} `json:"-"`

	RecvName string
	ArgNames []string
	ResNames []string

	FirstArgCtx   bool
	LastResultErr bool
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
		panic("unrecognized kind")
	}
}

func (c *FuncInfo) DisplayName() string {
	if c.RecvType != "" {
		return c.RecvType + "." + c.Name
	}
	return c.Name
}

type Object interface {
	GetField(name string) Field
	GetFieldIndex(i int) Field
	NumField() int
}

type ObjectWithErr interface {
	Object

	GetErr() Field
}

type Field interface {
	Name() string
	Value() interface{}
	Ptr() interface{}
	Set(val interface{})
}
