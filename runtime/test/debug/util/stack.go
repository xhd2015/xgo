package util

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

func BriefStack(stk *stack_model.Stack) string {
	var list []string
	for _, child := range stk.Children {
		list = append(list, briefStack(0, child))
	}
	return strings.Join(list, "\n")
}

func GetCostNs(stk *stack_model.Stack, caller string, callee string) int64 {
	callerFn := findFunc(stk.Children, caller)
	if callerFn == nil {
		return -1
	}
	calleeFn := findFunc(callerFn.Children, callee)
	if calleeFn == nil {
		return -1
	}
	return calleeFn.EndNs - calleeFn.BeginNs
}

func findFunc(stks []*stack_model.StackEntry, name string) *stack_model.StackEntry {
	for _, stk := range stks {
		fn := findFuncEntry(stk, name)
		if fn != nil {
			return fn
		}
	}
	return nil
}
func findFuncEntry(stk *stack_model.StackEntry, name string) *stack_model.StackEntry {
	if stk.FuncInfo != nil && stk.FuncInfo.Name == name {
		return stk
	}
	for _, child := range stk.Children {
		fn := findFuncEntry(child, name)
		if fn != nil {
			return fn
		}
	}
	return nil
}

func briefStack(indent int, stk *stack_model.StackEntry) string {
	var list []string
	var funcName string
	if stk.FuncInfo != nil {
		funcName = stk.FuncInfo.Name
	} else {
		funcName = "unknown"
	}
	list = append(list, fmt.Sprintf("%s%s", strings.Repeat(" ", indent), funcName))
	for _, child := range stk.Children {
		list = append(list, briefStack(indent+1, child))
	}
	return strings.Join(list, "\n")
}
