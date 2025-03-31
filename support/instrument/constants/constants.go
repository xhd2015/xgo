package constants

const (
	RUNTIME_TRAP_PKG             = "github.com/xhd2015/xgo/runtime/trap"
	RUNTIME_INTERNAL_RUNTIME_PKG = "github.com/xhd2015/xgo/runtime/internal/runtime"
	RUNTIME_CORE_PKG             = "github.com/xhd2015/xgo/runtime/core"
	RUNTIME_TRACE_SIGNAL_PKG     = "github.com/xhd2015/xgo/runtime/trace/signal"
	RUNTIME_TRAP_FLAGS_PKG       = "github.com/xhd2015/xgo/runtime/trap/flags"
	RUNTIME_FUNCTAB_PKG          = "github.com/xhd2015/xgo/runtime/functab"
)

const (
	RUNTIME_PKG_NAME_FUNC_TRAP = "__xgo_trap_runtime"
	RUNTIME_PKG_NAME_VAR_TRAP  = "__xgo_var_runtime"

	RUNTIME_PKG_NAME_FUNCTAB  = "__xgo_functab"
	RUNTIME_REGISTER_FUNC_TAB = "RegisterFunc"
)

const (
	RUNTIME_LINK_FILE = "runtime_link.go"
	VERSION_FILE      = "version.go"
	FLAG_FILE         = "flags.go"
)
