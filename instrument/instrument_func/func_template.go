package instrument_func

// funcInfo, recvPtr, paramPtrs, resultPtrs
const trapTemplate = "__xgo_post_%d, __xgo_stop_%d := __xgo_trap_%d(%s,%s,[]interface{}{%s},[]interface{}{%s});if __xgo_post_%d!=nil { defer __xgo_post_%d(); };if __xgo_stop_%d { return; };"
