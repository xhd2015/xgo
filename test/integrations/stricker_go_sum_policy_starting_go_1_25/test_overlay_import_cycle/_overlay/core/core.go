package core;import _ "github.com/xhd2015/xgo/test/integrations/stricker_go_sum_policy_starting_go_1_25/test_overlay_import_cycle/trace"

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
