package instrument_func

import "github.com/xhd2015/xgo/support/instrument/constants"

const trapTemplate = "__xgo_post_%d, __xgo_stop_%d := " + constants.RUNTIME_PKG_NAME_FUNC_TRAP + ".XgoTrap(%s,%s,[]string{%s},[]interface{}{%s},[]string{%s},[]interface{}{%s});if __xgo_post_%d!=nil { defer __xgo_post_%d(); };if __xgo_stop_%d { return; };"
