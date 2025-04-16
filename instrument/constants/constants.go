package constants

import "path/filepath"

const RUNTIME_MODULE = "github.com/xhd2015/xgo/runtime"

const (
	RUNTIME_INTERNAL_RUNTIME_PKG = "github.com/xhd2015/xgo/runtime/internal/runtime"
	RUNTIME_INTERNAL_TRAP_PKG    = "github.com/xhd2015/xgo/runtime/internal/trap"
	RUNTIME_TRAP_FLAGS_PKG       = "github.com/xhd2015/xgo/runtime/internal/flags"
	RUNTIME_FUNCTAB_PKG          = "github.com/xhd2015/xgo/runtime/functab"
	RUNTIME_CORE_PKG             = "github.com/xhd2015/xgo/runtime/core"
	RUNTIME_MOCK_PKG             = "github.com/xhd2015/xgo/runtime/mock"
	RUNTIME_TRACE_PKG            = "github.com/xhd2015/xgo/runtime/trace"
	RUNTIME_TRAP_PKG             = "github.com/xhd2015/xgo/runtime/trap"
)

// legacy
const (
	// Deprecated: we can remove once xgo/runtime v1.1.0 no longer used
	RUNTIME_LEGACY_CORE_INFO_PKG = "github.com/xhd2015/xgo/runtime/core/info"
)

const (
	LINK_REGISTER     = "__xgo_register_"
	LINK_TRAP_FUNC    = "__xgo_trap_"
	LINK_TRAP_VAR     = "__xgo_trap_var_"
	LINK_TRAP_VAR_PTR = "__xgo_trap_varptr_"

	RUNTIME_FUNCTAB_REGISTER = "RegisterFunc"
)

const (
	RUNTIME_LINK_TEMPLATE_FILE = "runtime_link_template.go"
	RUNTIME_LINK_FILE          = "runtime_link.go" // xgo/runtime/internal/runtime/runtime_link.go

	XGO_TRAP_TEMPLATE_FILE = "xgo_trap_template.go"
	XGO_TRAP_FILE          = "xgo_trap.go" // GOROOT/src/runtime/xgo_trap.go

	FUNCTAB_FILE = "functab.go"

	VERSION_FILE = "version.go"
	FLAG_FILE    = "flags.go"
	TRACE_FILE   = "trace.go"
)

const (
	FUNC_INFO = "__xgo_func_info"
	VAR_INFO  = "__xgo_var_info"
	INTF_INFO = "__xgo_intf_info"
)

const (
	XGO_REAL_NOW   = "XgoRealNow"
	XGO_REAL_SLEEP = "XgoRealSleep"
)

const (
	// see https://github.com/xhd2015/xgo/blob/branch-xgo-v1.0/runtime/core/version.go
	// the corresponding commit is 4123ef9cd711daea863cd3cf319989a581debaad
	LATEST_LEGACY_RUNTIME_NUMBER = 324
)

var RUNTIME_LINK_TEMPLATE_PATH = []string{"internal", "runtime", RUNTIME_LINK_TEMPLATE_FILE}
var RUNTIME_XGO_TRAP_TEMPLATE_PATH = []string{"internal", "runtime", XGO_TRAP_TEMPLATE_FILE}
var __GO_RUNTIME_XGO_TRAP_PATH = []string{"src", "runtime", XGO_TRAP_FILE}

func GetGoRuntimeXgoTrapFile(goroot string) string {
	return filepath.Join(goroot, filepath.Join(__GO_RUNTIME_XGO_TRAP_PATH...))
}
