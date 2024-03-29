package ctxt

import "strings"

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
