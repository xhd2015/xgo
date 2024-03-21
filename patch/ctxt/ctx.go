package ctxt

import (
	"cmd/compile/internal/base"
	"strings"
)

const XgoModule = "github.com/xhd2015/xgo"
const XgoRuntimePkg = XgoModule + "/runtime"

func SkipPackageTrap() bool {
	if base.Flag.Std {
		return true
	}

	pkgPath := GetPkgPath()
	if IsPkgXgoSkipTrap(pkgPath) {
		return true
	}
	// debug
	if strings.HasPrefix(pkgPath, "crypto/") {
		return true
	}

	// TODO: may allow customize package filter
	return false
}

func IsPkgXgoSkipTrap(pkg string) bool {
	suffix, ok := cutPkgPrefix(pkg, XgoModule)
	if !ok {
		return false
	}
	if suffix == "" {
		return true
	}
	// check if the package is test, runtime/test
	_, ok = cutPkgPrefix(suffix, "test")
	if ok {
		return false
	}
	_, ok = cutPkgPrefix(suffix, "runtime/test")
	if ok {
		return false
	}
	return true
}

func cutPkgPrefix(s string, pkg string) (suffix string, ok bool) {
	if !strings.HasPrefix(s, pkg) {
		return "", false
	}
	if len(s) == len(pkg) {
		return "", true
	}
	n := len(pkg)
	if s[n] != '/' {
		return "", false
	}
	return s[n+1:], true
}
