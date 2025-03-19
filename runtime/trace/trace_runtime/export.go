package trace_runtime

import (
	"strings"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trace/trace_runtime/stack_model"
)

func ExportStack(stack *Stack) *stack_model.Stack {
	return &stack_model.Stack{
		Format:   "stack",
		Begin:    stack.Begin.Format(time.RFC3339),
		Children: ExportStackEntries(stack.Roots),
	}
}

func ExportStackEntries(entries []*StackEntry) []*stack_model.StackEntry {
	if entries == nil {
		return nil
	}
	list := make([]*stack_model.StackEntry, len(entries))
	for i, entry := range entries {
		list[i] = ExportStackEntry(entry)
	}
	return list
}

func ExportStackEntry(entry *StackEntry) *stack_model.StackEntry {
	if entry == nil {
		return nil
	}
	return &stack_model.StackEntry{
		FuncInfo: ExportFuncInfo(entry),
		BeginNs:  entry.StartNs,
		EndNs:    entry.EndNs,
		Args:     entry.Args,
		Results:  entry.Results,
		Panic:    entry.Panic,
		Error:    entry.Error,
		Children: ExportStackEntries(entry.Children),
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
