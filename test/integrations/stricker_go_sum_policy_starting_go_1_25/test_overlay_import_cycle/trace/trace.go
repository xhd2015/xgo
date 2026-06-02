package trace

import "github.com/xhd2015/xgo/test/integrations/stricker_go_sum_policy_starting_go_1_25/test_overlay_import_cycle/core"

type Config struct {
	FilterTrace func(funcInfo *core.FuncInfo) bool
}
