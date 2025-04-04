package instrument_func

import "github.com/xhd2015/xgo/instrument/constants"

// funcInfo, recvPtr, paramPtrs, resultPtrs
const trapTemplate = "__xgo_post_%d, __xgo_stop_%d := " + constants.RUNTIME_PKG_NAME_FUNC + ".XgoTrap(" + constants.UNSAFE_PKG_NAME_FUNC + ".Pointer(%s),%s,[]interface{}{%s},[]interface{}{%s});if __xgo_post_%d!=nil { defer __xgo_post_%d(); };if __xgo_stop_%d { return; };"
