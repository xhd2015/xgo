package core

import (
	"reflect"
)

const __XGO_SKIP_TRAP = true

type FuncInfo struct {
	// FullName string
	Pkg          string
	IdentityName string
	Name         string
	RecvType     string
	RecvPtr      bool

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
