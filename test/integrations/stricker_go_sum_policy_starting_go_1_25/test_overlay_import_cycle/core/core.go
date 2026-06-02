package core

type Kind int

const (
	Kind_Func   Kind = 0
	Kind_Var    Kind = 1
	Kind_VarPtr Kind = 2
	Kind_Const  Kind = 3
)

type FuncInfo struct {
	Kind     Kind
	FullName string
}
