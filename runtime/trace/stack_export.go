package trace

import "time"

type FuncKind string

const (
	FuncKind_Func   FuncKind = "func"
	FuncKind_Var    FuncKind = "var"
	FuncKind_VarPtr FuncKind = "var_ptr"
	FuncKind_Const  FuncKind = "const"
)

type RootExport struct {
	// current executed function
	Begin    time.Time
	Children []*StackExport
}

type StackExport struct {
	FuncInfo *FuncInfoExport

	Begin int64 // us
	End   int64 // us

	Args    interface{}
	Results interface{}

	// is recorded as snapshot
	Snapshot bool

	Panic bool
	Error string

	Children []*StackExport
}

type FuncInfoExport struct {
	// FullName string
	Kind         FuncKind
	Pkg          string
	IdentityName string
	Name         string
	RecvType     string
	RecvPtr      bool

	// interface method?
	Interface bool
	Generic   bool
	Closure   bool
	Stdlib    bool

	File string
	Line int

	RecvName string
	ArgNames []string
	ResNames []string

	// is first argument ctx
	FirstArgCtx bool
	// last last result error
	LastResultErr bool
}
