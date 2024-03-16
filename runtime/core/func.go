package core

import (
	"reflect"
	"strings"
)

const __XGO_SKIP_TRAP = true

type FuncInfo struct {
	FullName string
	Pkg      string
	RecvType string
	RecvPtr  bool
	Name     string

	Generic bool

	PC       uintptr     `json:"-"`
	Func     interface{} `json:"-"`
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

func (c *FuncInfo) IsFunc(fn interface{}) bool {
	if c.PC == 0 || fn == nil {
		return false
	}
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return false
	}
	return c.PC == v.Pointer()
}

// a/b/c.A
// a/b/c.(*C).X
// a/b/c.C.Y
// a/b/c.Z
func ParseFuncName(fullName string, hasPkg bool) (pkgPath string, recvName string, recvPtr bool, funcName string) {
	s := fullName
	funcNameDot := strings.LastIndex(s, ".")
	if funcNameDot < 0 {
		funcName = s
		return
	}
	funcName = s[funcNameDot+1:]
	s = s[:funcNameDot]

	recvName = s
	if hasPkg {
		recvDot := strings.LastIndex(s, ".")
		if recvDot < 0 {
			pkgPath = s
			return
		}
		recvName = s[recvDot+1:]
		s = s[:recvDot]
	}

	recvName = strings.TrimPrefix(recvName, "(")
	recvName = strings.TrimSuffix(recvName, ")")
	if strings.HasPrefix(recvName, "*") {
		recvPtr = true
		recvName = recvName[1:]
	}
	pkgPath = s

	return
}
