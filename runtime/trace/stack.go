package trace

import (
	"time"

	"github.com/xhd2015/xgo/runtime/core"
)

type Root struct {
	// current executed function
	Top      *Stack
	Begin    time.Time
	Children []*Stack
}

type Stack struct {
	FuncInfo *core.FuncInfo

	Begin int64 // us
	End   int64 // us

	Args    core.Object
	Results core.Object
	Panic   bool
	Error   error
	// Recv     interface{}
	// Args     []interface{}
	// Results  []interface{}
	Children []*Stack
}

func (c *Root) Export() *RootExport {
	if c == nil {
		return nil
	}
	return &RootExport{
		Begin:    c.Begin,
		Children: (stacks)(c.Children).Export(),
	}
}

type stacks []*Stack

func (c stacks) Export() []*StackExport {
	if c == nil {
		return nil
	}
	list := make([]*StackExport, len(c))
	for i := 0; i < len(c); i++ {
		list[i] = c[i].Export()
	}
	return list
}

func (c *Stack) Export() *StackExport {
	if c == nil {
		return nil
	}
	var errMsg string
	if c.Error != nil {
		errMsg = c.Error.Error()
	}
	return &StackExport{
		FuncInfo: ExportFuncInfo(c.FuncInfo),
		Begin:    c.Begin,
		End:      c.End,
		Args:     c.Args,
		Results:  c.Results,
		Panic:    c.Panic,
		Error:    errMsg,
		Children: (stacks)(c.Children).Export(),
	}
}

func ExportFuncInfo(c *core.FuncInfo) *FuncInfoExport {
	if c == nil {
		return nil
	}
	return &FuncInfoExport{
		Pkg:          c.Pkg,
		IdentityName: c.IdentityName,
		Name:         c.Name,
		RecvType:     c.RecvType,
		RecvPtr:      c.RecvPtr,

		Generic:  c.Generic,
		RecvName: c.RecvName,
		ArgNames: c.ArgNames,
		ResNames: c.ResNames,

		FirstArgCtx:   c.FirstArgCtx,
		LastResultErr: c.LastResultErr,
	}
}
