package stack_model

// keep the same as cmd/xgo/trace/render/stack_model.go

type FuncKind string

const (
	FuncKind_Func   FuncKind = "func"
	FuncKind_Var    FuncKind = "var"
	FuncKind_VarPtr FuncKind = "var_ptr"
	FuncKind_Const  FuncKind = "const"
)

type Stack struct {
	Format   string
	Begin    string
	Children []*StackEntry
}

type StackEntry struct {
	FuncInfo *FuncInfo

	BeginNs int64 // ns
	EndNs   int64 // ns

	Args    interface{}
	Results interface{}

	Panic bool
	Error string

	Children []*StackEntry
}

type FuncInfo struct {
	// FullName string
	Kind     FuncKind
	Pkg      string
	Name     string
	RecvType string
	RecvPtr  bool

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
