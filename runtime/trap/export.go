package trap

import (
	"strings"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap/stack_model"
)

func ExportStack(stack *Stack, offsetNS int64) *stack_model.Stack {
	if stack == nil {
		return nil
	}
	return &stack_model.Stack{
		Format:   "stack",
		Begin:    stack.Begin.Format(time.RFC3339),
		Children: ExportStackEntries(stack.Roots, offsetNS),
	}
}

func ExportStackEntries(entries []*StackEntry, offsetNS int64) []*stack_model.StackEntry {
	if entries == nil {
		return nil
	}
	list := make([]*stack_model.StackEntry, len(entries))
	for i, entry := range entries {
		list[i] = ExportStackEntry(entry, offsetNS)
	}
	return list
}

func ExportStackEntry(entry *StackEntry, offsetNS int64) *stack_model.StackEntry {
	if entry == nil {
		return nil
	}

	children := ExportStackEntries(entry.Children, offsetNS)
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
			exportedStack := ExportStack(stack, beginNs)
			children = append(children, exportedStack.Children...)
			if stack.End.IsZero() {
				fnInfo.Name += " (running)"
				endNs = time.Now().UnixNano() - stack.Begin.UnixNano() + beginNs
			} else {
				endNs = stack.End.UnixNano() - stack.Begin.UnixNano() + beginNs
			}
		}
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

func ExportFuncInfo(entry *StackEntry) *stack_model.FuncInfo {
	if entry == nil {
		return nil
	}
	pkg, name := splitFuncPkg(entry.FuncName)
	return &stack_model.FuncInfo{
		Name: name,
		Pkg:  pkg,
		File: entry.File,
		Line: entry.Line,
	}
}

func splitFuncPkg(funcName string) (string, string) {
	pkg, recvName, recvPtr, typeGeneric, funcGeneric, basicName := core.ParseFuncName(funcName)

	_ = recvName
	_ = recvPtr
	_ = typeGeneric
	_ = funcGeneric
	_ = basicName

	if pkg == "" {
		return "", funcName
	}
	name := strings.TrimPrefix(funcName, pkg)
	name = strings.TrimPrefix(name, ".")

	return pkg, name
}
