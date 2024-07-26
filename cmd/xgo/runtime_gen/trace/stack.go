package trace

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
)

// default size to shrink 16K
const DefaultSizeLimit = 16 * 1024

// default appearance limit on repeative functions
const DefaultAppearanceLimit = 100

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

	// is recorded as snapshot
	Snapshot bool

	Panic    bool
	Error    error
	Children []*Stack
}

// allow skip some packages
//
//	for example: google.golang.org/protobuf/internal/order
type ExportOptions struct {
	// suppress error when marshalling
	// arguments and results
	DisableErrSilent bool

	SizeLimit       int // 0: default limit 16K
	AppearanceLimit int // 0: default limit 100

	FilterStack func(stack *StackExport) *StackExport

	FilterRoot  func(root *RootExport) *RootExport
	MarshalRoot func(root *RootExport) ([]byte, error)

	stats map[string]map[string]*stat
}
type stat struct {
	total   int
	current int
}

func (c *ExportOptions) getSizeLimit() int {
	if c == nil || c.SizeLimit == 0 {
		return DefaultSizeLimit
	}
	return c.SizeLimit
}
func (c *ExportOptions) getAppearanceLimit() int {
	if c == nil || c.AppearanceLimit == 0 {
		return DefaultAppearanceLimit
	}
	return c.AppearanceLimit
}

func (c *Root) Export(opts *ExportOptions) *RootExport {
	if c == nil {
		return nil
	}
	if opts == nil {
		opts = &ExportOptions{}
	}
	if opts.getAppearanceLimit() > 0 {
		opts.stats = getStats(c)
	}

	return &RootExport{
		Begin:    c.Begin,
		Children: (stacks)(c.Children).Export(opts),
	}
}

func getStats(root *Root) map[string]map[string]*stat {
	mapping := make(map[string]map[string]*stat)
	var traverse func(stack *Stack)
	traverse = func(st *Stack) {
		if st == nil {
			return
		}
		if st.FuncInfo != nil {
			pkg := st.FuncInfo.Pkg
			fn := st.FuncInfo.IdentityName
			fnMapping := mapping[pkg]
			if fnMapping == nil {
				fnMapping = make(map[string]*stat)
				mapping[pkg] = fnMapping
			}
			st := fnMapping[fn]
			if st == nil {
				st = &stat{}
				fnMapping[fn] = st
			}
			st.total++
		}

		for _, st := range st.Children {
			traverse(st)
		}
	}
	for _, st := range root.Children {
		traverse(st)
	}
	return mapping
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
		if exportStack.FuncInfo != nil && opts != nil && opts.stats != nil {
			apprLimit := opts.getAppearanceLimit()
			if apprLimit > 0 {
				fnStat := opts.stats[exportStack.FuncInfo.Pkg][exportStack.FuncInfo.IdentityName]
				if fnStat != nil && fnStat.total > apprLimit {
					if fnStat.current >= apprLimit {
						continue
					}
					fnStat.current++
				}
			}
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
	var args interface{} = c.Args
	var results interface{} = c.Results

	sizeLimit := opts.getSizeLimit()
	if sizeLimit > 0 {
		args = &LimitSize{args, sizeLimit}
		results = &LimitSize{results, sizeLimit}
	}
	if opts == nil || !opts.DisableErrSilent {
		args = &ErrSilent{args}
		results = &ErrSilent{results}
	}
	stack := &StackExport{
		FuncInfo: ExportFuncInfo(c.FuncInfo, opts),
		Begin:    c.Begin,
		End:      c.End,
		Args:     args,
		Results:  results,
		Snapshot: c.Snapshot,
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
		Kind:         FuncKind(c.Kind.String()),
		Pkg:          c.Pkg,
		IdentityName: c.IdentityName,
		Name:         c.Name,
		RecvType:     c.RecvType,
		RecvPtr:      c.RecvPtr,

		Interface: c.Interface,
		Generic:   c.Generic,
		Closure:   c.Closure,
		Stdlib:    c.Stdlib,
		RecvName:  c.RecvName,
		ArgNames:  c.ArgNames,
		ResNames:  c.ResNames,

		FirstArgCtx:   c.FirstArgCtx,
		LastResultErr: c.LastResultErr,

		File: c.File,
		Line: c.Line,
	}
}

// make json err silent
type ErrSilent struct {
	Data interface{}
}

func (c *ErrSilent) MarshalJSON() (data []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
		if err != nil {
			data = []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
			err = nil
		}
	}()
	data, err = json.Marshal(c.Data)
	return
}

// make json err silent
type LimitSize struct {
	Data  interface{}
	Limit int
}

func (c *LimitSize) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(c.Data)
	if err != nil {
		return nil, err
	}
	if c.Limit <= 0 || c.Limit >= len(data) {
		return data, nil
	}
	// shorten
	return []byte(fmt.Sprintf(`{"size":%d, "sizeBeforeShrink":%d,"partialData":%q}`, c.Limit, len(data), string(data[:c.Limit]))), nil
}
