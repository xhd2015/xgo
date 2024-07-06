package ctxt

import (
	"cmd/compile/internal/xgo_rewrite_internal/patch/info"
	"cmd/compile/internal/xgo_rewrite_internal/patch/match"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

var opts Options

func init() {
	if XGO_COMPILER_OPTIONS_FILE == "" {
		return
	}
	data, err := ioutil.ReadFile(XGO_COMPILER_OPTIONS_FILE)
	if err != nil {
		panic(fmt.Errorf("read xgo compiler options: %w", err))
	}
	if len(data) == 0 {
		return
	}
	err = json.Unmarshal(data, &opts)
	if err != nil {
		panic(fmt.Errorf("parse xgo compiler options: %w", err))
	}
	n := len(opts.FilterRules)
	for i := 0; i < n; i++ {
		opts.FilterRules[i].Parse()
	}
}

func GetAction(fn *info.DeclInfo) string {
	return match.MatchRules(opts.FilterRules, GetPkgPath(), isMainModule, fn)
}

func MatchAnyPattern(pkgPath string, pkgName string, funcName string, patterns []string) bool {
	for _, pattern := range patterns {
		if MatchPattern(pkgPath, pkgName, funcName, pattern) {
			return true
		}
	}
	return false
}

func MatchPattern(pkgPath string, pkgName string, funcName string, pattern string) bool {
	dotIdx := strings.Index(pattern, ".")
	if dotIdx < 0 {
		return funcName == pattern
	}
	pkgPattern := pattern[:dotIdx]
	funcPattern := pattern[dotIdx+1:]

	var pkgMatch bool
	if pkgName != "" && pkgName == pkgPattern {
		pkgMatch = true
	} else {
		var pkgPathSuffix bool
		if strings.HasPrefix(pkgPattern, "*") {
			pkgPattern = pkgPattern[1:]
			pkgPathSuffix = true
		}
		if pkgPathSuffix {
			pkgMatch = pkgPattern == "" || strings.HasSuffix(pkgPath, pkgPattern)
		} else {
			pkgMatch = pkgPath == pkgPattern
		}
	}
	if !pkgMatch {
		return false
	}
	if funcPattern == "*" || funcPattern == funcName {
		return true
	}
	return false
}
