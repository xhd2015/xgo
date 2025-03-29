package instrument_func

const trapTemplate = "__xgo_post_%d, __xgo_stop_%d := __xgo_trap_runtime.XgoTrap(%s,%s,[]string{%s},[]interface{}{%s},[]string{%s},[]interface{}{%s});if __xgo_post_%d!=nil { defer __xgo_post_%d(); };if __xgo_stop_%d { return; };"
