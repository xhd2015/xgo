package ctxt

import "cmd/compile/internal/xgo_rewrite_internal/patch/match"

type Options struct {
	FilterRules []match.Rule `json:"filter_rules"`
}
