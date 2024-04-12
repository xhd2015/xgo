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

	Args     core.Object
	Results  core.Object
	Panic    bool
	Error    error
	Children []*Stack
}

// allow skip some packages
//
//	for example: google.golang.org/protobuf/internal/order
type ExportOptions struct {
	FilterStack func(stack *StackExport) *StackExport
}

func (c *Root) Export(opts *ExportOptions) *RootExport {
	if c == nil {
		return nil
	}
	return &RootExport{
		Begin:    c.Begin,
		Children: (stacks)(c.Children).Export(opts),
	}
}

type stacks []*Stack

func (c stacks) Export(opts *ExportOptions) []*StackExport {
	if c == nil {
		return nil
	}
	list := make([]*StackExport, 0, len(c))
	for i := 0; i < len(c); i++ {
		stackPoint := c[i]
		exportStack := stackPoint.Export(opts)
		if exportStack == nil {
			continue
		}
		list = append(list, exportStack)
	}
	return list
}

func (c *Stack) Export(opts *ExportOptions) *StackExport {
	if c == nil {
		return nil
	}
	var errMsg string
	if c.Error != nil {
		errMsg = c.Error.Error()
	}
	stack := &StackExport{
		FuncInfo: ExportFuncInfo(c.FuncInfo, opts),
		Begin:    c.Begin,
		End:      c.End,
		Args:     c.Args,
		Results:  c.Results,
		Panic:    c.Panic,
		Error:    errMsg,
		Children: ((stacks)(c.Children)).Export(opts),
	}

	if opts != nil && opts.FilterStack != nil {
		return opts.FilterStack(stack)
	}
	return stack
}

func ExportFuncInfo(c *core.FuncInfo, opts *ExportOptions) *FuncInfoExport {
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

		File: c.File,
		Line: c.Line,
	}
}
