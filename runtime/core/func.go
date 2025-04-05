package core

import (
	"strings"

	"github.com/xhd2015/xgo/runtime/core/info"
)

// core.Object, core.FuncInfo are used in interceptor signature

type Kind = info.Kind

const (
	Kind_Func   = info.Kind_Func
	Kind_Var    = info.Kind_Var
	Kind_VarPtr = info.Kind_VarPtr
	Kind_Const  = info.Kind_Const
)

type FuncInfo = info.Func

// a/b/c.A
// a/b/c.(*C).X
// a/b/c.C.Y
// a/b/c.Z
// a/b/c.Z[].X
// a/b/c.X[]
//
// parse process:
//
//	funcGeneric
//	funcName
//	recvType
//	recvGeneric
//	pkgPath
func ParseFuncName(fullName string) (pkgPath string, recvName string, recvPtr bool, typeGeneric string, funcGeneric string, funcName string) {
	sepIdx := strings.LastIndex(fullName, "/")
	pkg_recv_func_no_generic := fullName
	// func generic
	if strings.HasSuffix(pkg_recv_func_no_generic, "]") {
		leftIdx := strings.LastIndex(pkg_recv_func_no_generic, "[")
		if leftIdx < 0 {
			// invalid
			return
		}
		funcGeneric = pkg_recv_func_no_generic[leftIdx+1 : len(pkg_recv_func_no_generic)-1]
		pkg_recv_func_no_generic = pkg_recv_func_no_generic[:leftIdx]
	}

	// func name
	funcNameDot := strings.LastIndex(pkg_recv_func_no_generic, ".")
	if funcNameDot < 0 {
		funcName = pkg_recv_func_no_generic
		return
	}
	funcName = pkg_recv_func_no_generic[funcNameDot+1:]

	// Type
	pkg_recv_generic_paren := pkg_recv_func_no_generic[:funcNameDot]

	var hasParen bool
	pkg_recv_generic := pkg_recv_generic_paren
	if strings.HasSuffix(pkg_recv_generic_paren, ")") {
		// receiver wrap
		leftIdx := strings.LastIndex(pkg_recv_generic_paren, "(")
		if leftIdx < 0 {
			// invalid
			return
		}
		pkg_recv_generic = pkg_recv_generic_paren[leftIdx+1 : len(pkg_recv_generic_paren)-1]
		if pkg_recv_generic_paren[leftIdx-1] != '.' {
			// invalid
			return
		}
		pkgPath = pkg_recv_generic_paren[:leftIdx-1]
		hasParen = true
	}

	pkg_recv := pkg_recv_generic
	// parse generic
	if strings.HasSuffix(pkg_recv_generic, "]") {
		leftIdx := strings.LastIndex(pkg_recv_generic, "[")
		if leftIdx < 0 {
			// invalid
			return
		}
		typeGeneric = pkg_recv_generic[leftIdx+1 : len(pkg_recv_generic)-1]
		pkg_recv = pkg_recv_generic[:leftIdx]
	}

	// pkgPath_recv
	recv_ptr := pkg_recv
	if !hasParen {
		dotIdx := strings.LastIndex(pkg_recv, ".")
		if dotIdx < 0 {
			pkgPath = recv_ptr
			// invalid
			return
		}
		if dotIdx < sepIdx {
			if strings.Contains(typeGeneric, "/") {
				// recheck, see bug https://github.com/xhd2015/xgo/issues/211
				sepIdx = strings.LastIndex(pkg_recv, "/")
			}
			if dotIdx < sepIdx {
				// no recv
				pkgPath = recv_ptr
				return
			}
		}
		pkgPath = pkg_recv[:dotIdx]
		recv_ptr = pkg_recv[dotIdx+1:]
	}

	recvName = recv_ptr
	if strings.HasPrefix(recv_ptr, "*") {
		recvPtr = true
		recvName = recv_ptr[1:]
	}

	return
}
