package core

import (
	"fmt"
	"strings"
)

const __XGO_SKIP_TRAP = true

type Kind int

const (
	Kind_Func   Kind = 0
	Kind_Var    Kind = 1
	Kind_VarPtr Kind = 2
	Kind_Const  Kind = 3
)

func (c Kind) String() string {
	switch c {
	case Kind_Func:
		return "func"
	case Kind_Var:
		return "var"
	case Kind_VarPtr:
		return "var_ptr"
	case Kind_Const:
		return "const"
	default:
		return fmt.Sprintf("kind_%d", int(c))
	}
}

type FuncInfo struct {
	Kind Kind
	// full name, format: {pkgPath}.{receiver}.{funcName}
	// example:  github.com/xhd2015/xgo/runtime/core.(*SomeType).SomeFunc
	FullName string
	Pkg      string
	// identity name within a package, for ptr-method, it's something like `(*SomeType).SomeFunc`
	// run `go run ./test/example/method` to verify
	IdentityName string
	Name         string
	RecvType     string
	RecvPtr      bool

	// is this an interface method?
	Interface bool

	// is this a generic function?
	Generic bool

	// is this a closure?
	Closure bool

	// is this function from stdlib
	Stdlib bool

	// source info
	File string
	Line int

	PC   uintptr     `json:"-"`
	Func interface{} `json:"-"`
	Var  interface{} `json:"-"` // var address

	RecvName string
	ArgNames []string
	ResNames []string

	// is first argument ctx
	FirstArgCtx bool
	// last last result error
	LastResultErr bool
}

func (c *FuncInfo) DisplayName() string {
	if c.RecvType != "" {
		return c.RecvType + "." + c.Name
	}
	return c.Name
}

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
