package ctxt

import (
	"cmd/compile/internal/base"
	"strings"
)

const XgoModule = "github.com/xhd2015/xgo"
const XgoRuntimePkg = XgoModule + "/runtime"
const XgoRuntimeCorePkg = XgoModule + "/runtime/core"
const XgoRuntimeTracePkg = XgoModule + "/runtime/trace"

const XgoLinkTrapVarForGenerated = "__xgo_link_trap_var_for_generated"

func InitAfterLoad() {
	isMainModule = IsSameModule(GetPkgPath(), XgoMainModule)
}

func SkipPackageTrap() bool {
	pkgPath := GetPkgPath()
	if pkgPath == "" {
		return true
	}
	if pkgPath == "runtime" || strings.HasPrefix(pkgPath, "runtime/") || strings.HasPrefix(pkgPath, "internal/") {
		return true
	}
	if base.Flag.Std {
		// skip std lib, especially skip:
		//    runtime, runtime/internal, runtime/*, reflect, unsafe, syscall, sync, sync/atomic,  internal/*
		//
		// however, there are some funcs in stdlib that we can
		// trap, for example, db connection
		// for example:
		//     errors, math, math/bits, unicode, unicode/utf8, unicode/utf16, strconv, path, sort, time, encoding/json

		// NOTE: base.Flag.Std in does not always reflect func's package path,
		// because generic instantiation happens in other package, so this
		// func may be a foreigner.

		if XgoStdTrapDefaultAllow {
			if _, ok := stdBlocklist[pkgPath]["*"]; ok {
				return true
			}
			return false
		}
		// allow http
		if _, ok := stdWhitelist[pkgPath]; ok {
			return false
		}
		return true
	}
	if isSkippableSpecialPkg(pkgPath) {
		return true
	}

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

// always trap for encoding/json.newTypeEncoder
func AlwaysTrap(pkgPath string, isStd bool, identityName string, funcName string) bool {
	if !isStd {
		return false
	}
	if pkgPath == "encoding/json" && identityName == "newTypeEncoder" {
		return true
	}
	return false
}

func AllowPkgFuncTrap(pkgPath string, isStd bool, identityName string, funcName string) bool {
	if isStd {
		return allowStdFunc(pkgPath, identityName, funcName)
	}

	return true
}

// skip all packages for xgo,except test
func IsPkgXgoSkipTrap(pkg string) bool {
	suffix, ok := cutPkgPrefix(pkg, XgoModule)
	if !ok {
		return false
	}
	if suffix == "" {
		return true
	}
	// check if the package is test or runtime/test
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

var isMainModule bool

func IsMainModule() bool {
	return isMainModule
}

func IsPkgMainModule(pkg string) bool {
	return IsSameModule(pkg, XgoMainModule)
}

func IsSameModule(pkgPath string, modulePath string) bool {
	if modulePath == "" {
		return false
	}
	if !strings.HasPrefix(pkgPath, modulePath) {
		return false
	}
	if len(pkgPath) == len(modulePath) {
		return true
	}
	if pkgPath[len(modulePath)] == '/' {
		return true
	}
	return false
}
