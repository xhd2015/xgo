package ctxt

import "os"

var XgoMainModule = os.Getenv("XGO_MAIN_MODULE")
var XgoCompilePkgDataDir = os.Getenv("XGO_COMPILE_PKG_DATA_DIR")

// enabled via: --trap-stdlib
var XgoStdTrapDefaultAllow = os.Getenv("XGO_STD_LIB_TRAP_DEFAULT_ALLOW") == "true"
