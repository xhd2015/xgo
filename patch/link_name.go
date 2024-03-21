package patch

import (
	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
)

const xgoRuntimePkgPrefix = xgo_ctxt.XgoRuntimePkg + "/"
const xgoRuntimeTrapPkg = xgoRuntimePkgPrefix + "trap"

// accepts interface{} as argument
const xgoOnTestStart = "__xgo_on_test_start"

const setTrap = "__xgo_set_trap"

var linkMap = map[string]string{
	"__xgo_link_getcurg":                      "__xgo_getcurg",
	"__xgo_link_set_trap":                     setTrap,
	"__xgo_link_init_finished":                "__xgo_init_finished",
	"__xgo_link_on_init_finished":             "__xgo_on_init_finished",
	"__xgo_link_on_goexit":                    "__xgo_on_goexit",
	"__xgo_link_on_test_start":                xgoOnTestStart,
	"__xgo_link_get_test_starts":              "__xgo_get_test_starts",
	"__xgo_link_retrieve_all_funcs_and_clear": "__xgo_retrieve_all_funcs_and_clear",
	"__xgo_link_peek_panic":                   "__xgo_peek_panic",
}
