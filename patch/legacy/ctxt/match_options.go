package ctxt

import "cmd/compile/internal/xgo_rewrite_internal/patch/match"

// Options is read from XGO_COMPILER_OPTIONS_FILE (options-from-file.json).
// Keep in sync with cmd/xgo.FileOptions.
type Options struct {
	FilterRules []match.Rule `json:"filter_rules"`
	// MockRuleIncludeAsMainModule: additive module paths for main_module rules.
	MockRuleIncludeAsMainModule []string `json:"mock_rule_include_as_main_module,omitempty"`
}
