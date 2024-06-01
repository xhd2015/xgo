package match

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
	"cmd/compile/internal/xgo_rewrite_internal/patch/info"
	"strings"
)

type Rule struct {
	Any        bool    `json:"any"`
	Kind       *string `json:"kind"`
	Pkg        *string `json:"pkg"`
	Name       *string `json:"name"`
	Stdlib     *bool   `json:"stdlib"`
	MainModule *bool   `json:"main_module"`
	Generic    *bool   `json:"generic"`
	Exported   *bool   `json:"exported"`
	Action     string  `json:"action"` // include,exclude or empty

	kinds []string
	pkgs  Patterns
	names Patterns
}

func (c *Rule) Parse() {
	c.kinds = toList(c.Kind)
	c.pkgs = CompilePatterns(toList(c.Pkg))
	c.names = CompilePatterns(toList(c.Name))
}

func toList(s *string) []string {
	if s == nil {
		return nil
	}
	list := strings.Split(*s, ",")
	i := 0
	n := len(list)
	for j := 0; j < n; j++ {
		e := list[j]
		e = strings.TrimSpace(e)
		if e != "" {
			list[i] = e
			i++
		}
	}
	return list[:i]
}

func MatchRules(rules []Rule, pkgPath string, isMainModule bool, funcDecl *info.DeclInfo) string {
	for _, rule := range rules {
		if Match(&rule, pkgPath, isMainModule, funcDecl) {
			return rule.Action
		}
	}
	return ""
}

func Match(rule *Rule, pkgPath string, isMainModule bool, funcDecl *info.DeclInfo) bool {
	if rule == nil {
		return false
	}
	if rule.Any {
		return true
	}
	var hasAnyCondition bool
	if len(rule.kinds) > 0 {
		hasAnyCondition = true
		if !listContains(rule.kinds, funcDecl.Kind.String()) {
			return false
		}
	}
	if len(rule.pkgs) > 0 {
		hasAnyCondition = true
		if !rule.pkgs.MatchAny(pkgPath) {
			return false
		}
	}
	if len(rule.names) > 0 {
		hasAnyCondition = true
		if !rule.names.MatchAny(funcDecl.IdentityName()) {
			return false
		}
	}
	if rule.MainModule != nil {
		hasAnyCondition = true
		if *rule.MainModule != isMainModule {
			return false
		}
	}
	if rule.Stdlib != nil {
		hasAnyCondition = true
		if *rule.Stdlib != base.Flag.Std {
			return false
		}
	}
	if rule.Generic != nil && funcDecl.Kind.IsFunc() {
		hasAnyCondition = true
		if *rule.Generic != funcDecl.Generic {
			return false
		}
	}
	if rule.Exported != nil {
		hasAnyCondition = true
		if *rule.Exported != types.IsExported(funcDecl.Name) {
			return false
		}
	}
	if hasAnyCondition {
		return true
	}
	return false
}

func listContains(list []string, e string) bool {
	for _, x := range list {
		if x == e {
			return true
		}
	}
	return false
}
