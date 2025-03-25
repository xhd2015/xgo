package inject

import (
	_ "embed"

	"github.com/xhd2015/xgo/support/instrument/patch"
)

const XGO_RUNTIME_PKG = "github.com/xhd2015/xgo/runtime/trace/trace_runtime"

const LINK_FILE = "runtime_link.go"

//go:embed runtime_link_template.go
var runtimeLinkTemplate string

func GetLinkRuntimeCode() string {
	code, err := patch.RemoveBuildIgnore([]byte(runtimeLinkTemplate))
	if err != nil {
		panic(err)
	}
	return string(code)
}
