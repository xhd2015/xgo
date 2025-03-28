package instrument_xgo_runtime

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/instrument/patch"
)

const (
	RUNTIME_TRAP_PKG             = "github.com/xhd2015/xgo/runtime/trap"
	RUNTIME_INTERNAL_RUNTIME_PKG = "github.com/xhd2015/xgo/runtime/internal/runtime"
	RUNTIME_CORE_PKG             = "github.com/xhd2015/xgo/runtime/core"
	RUNTIME_TRACE_SIGNAL_PKG     = "github.com/xhd2015/xgo/runtime/trace/signal"
	RUNTIME_TRAP_FLAGS_PKG       = "github.com/xhd2015/xgo/runtime/trap/flags"
)

const (
	RUNTIME_LINK_FILE = "runtime_link.go"
	VERSION_FILE      = "version.go"
	FLAG_FILE         = "flags.go"
)

//go:embed runtime_link_template.go
var runtimeLinkTemplate string

func GetLinkRuntimeCode() string {
	code, err := patch.RemoveBuildIgnore(runtimeLinkTemplate)
	if err != nil {
		panic(err)
	}
	return code
}

func ReplaceVersion(versionCode string, xgoVersion string, xgoRevision string, xgoNumber int) string {
	versionCode = strings.ReplaceAll(versionCode, `const XGO_VERSION = ""`, `const XGO_VERSION = "`+xgoVersion+`"`)
	versionCode = strings.ReplaceAll(versionCode, `const XGO_REVISION = ""`, `const XGO_REVISION = "`+xgoRevision+`"`)
	versionCode = strings.ReplaceAll(versionCode, `const XGO_NUMBER = 0`, `const XGO_NUMBER = `+strconv.Itoa(xgoNumber))
	return versionCode
}

func InjectFlags(flagsCode string, collectTestTrace bool, collectTestTraceDir string) string {
	flagsCode = strings.ReplaceAll(flagsCode, `const COLLECT_TEST_TRACE = false`, fmt.Sprintf(`const COLLECT_TEST_TRACE = %t`, collectTestTrace))
	flagsCode = strings.ReplaceAll(flagsCode, `const COLLECT_TEST_TRACE_DIR = ""`, fmt.Sprintf(`const COLLECT_TEST_TRACE_DIR = %q`, collectTestTraceDir))
	return flagsCode
}
