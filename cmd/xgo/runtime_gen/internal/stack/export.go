package stack

import (
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

func Export(stack *Stack, offsetNS int64) *stack_model.Stack {
	if stack == nil {
		return nil
	}
	return &stack_model.Stack{
		Format:   "stack",
		Begin:    stack.Begin.Format(time.RFC3339),
		Children: ExportStackEntries(stack.Roots, stack.Begin, offsetNS),
	}
}

func ExportStackEntries(entries []*Entry, rootBegin time.Time, offsetNS int64) []*stack_model.StackEntry {
	if entries == nil {
		return nil
	}
	list := make([]*stack_model.StackEntry, len(entries))
	for i, entry := range entries {
		list[i] = ExportStackEntry(entry, rootBegin, offsetNS)
	}
	return list
}

func ExportStackEntry(entry *Entry, rootBegin time.Time, offsetNS int64) *stack_model.StackEntry {
	if entry == nil {
		return nil
	}

	var isRunning bool
	children := ExportStackEntries(entry.Children, rootBegin, offsetNS)
	beginNs := entry.BeginNs + offsetNS
	// stackEndNs := entry.EndNs
	endNs := entry.EndNs + offsetNS
	fnInfo := ExportFuncInfo(entry)
	if entry.Go && entry.GetStack != nil {
		// handle async stack
		stack := entry.GetStack()
		if stack != nil {
			// NOTE: this might be unsafe since the
			// child goroutine might be running
			exportedStack := Export(stack, beginNs)
			children = append(children, exportedStack.Children...)
			if stack.End.IsZero() {
				isRunning = true
				endNs = runtime.XgoRealTimeNow().UnixNano() - stack.Begin.UnixNano() + beginNs
			} else {
				endNs = stack.End.UnixNano() - stack.Begin.UnixNano() + beginNs
			}
		}
	} else if !entry.Finished {
		endNs += runtime.XgoRealTimeNow().UnixNano() - rootBegin.UnixNano()
		isRunning = true
	}
	if isRunning {
		fnInfo.Name += " (running)"
	}
	return &stack_model.StackEntry{
		FuncInfo: fnInfo,
		BeginNs:  beginNs,
		EndNs:    endNs,
		Args:     entry.Args,
		Results:  entry.Results,
		Panic:    entry.Panic,
		Error:    entry.Error,
		Children: children,
	}
}

func ExportFuncInfo(entry *Entry) *stack_model.FuncInfo {
	if entry == nil {
		return nil
	}
	// the file and line is where the call occurs
	// not where the variable is defined
	file := entry.File
	line := entry.Line

	var pkg string
	var name string
	var kind stack_model.FuncKind

	var recvName string
	var argNames []string
	var resNames []string

	var generic bool

	var firstArgCtx bool
	var lastResultErr bool

	name = entry.FuncName
	if entry.FuncInfo != nil {
		pkg = entry.FuncInfo.Pkg
		name = entry.FuncInfo.IdentityName

		recvName = entry.FuncInfo.RecvName
		argNames = entry.FuncInfo.ArgNames
		resNames = entry.FuncInfo.ResNames

		generic = entry.FuncInfo.Generic

		firstArgCtx = entry.FuncInfo.FirstArgCtx
		lastResultErr = entry.FuncInfo.LastResultErr

		switch entry.FuncInfo.Kind {
		case core.Kind_Func:
			kind = stack_model.FuncKind_Func
			file = entry.FuncInfo.File
			line = entry.FuncInfo.Line
		case core.Kind_Var:
			kind = stack_model.FuncKind_Var
		case core.Kind_VarPtr:
			kind = stack_model.FuncKind_VarPtr
		case core.Kind_Const:
			kind = stack_model.FuncKind_Const
		}
	}

	return &stack_model.FuncInfo{
		Kind:          kind,
		Name:          name,
		Pkg:           pkg,
		File:          file,
		Line:          line,
		RecvName:      recvName,
		ArgNames:      argNames,
		ResNames:      resNames,
		FirstArgCtx:   firstArgCtx,
		LastResultErr: lastResultErr,
		Generic:       generic,
	}
}
