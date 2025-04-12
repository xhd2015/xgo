package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/build/modules_without_trap_all/service/pkg"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

func TestTrace(t *testing.T) {
	var stack stack_model.IStack
	trace.Trace(trace.Config{
		OnFinish: func(s stack_model.IStack) {
			stack = s
		},
	}, nil, func() (interface{}, error) {
		run()
		return nil, nil
	})
	// only main module is traced without --trap-all
	expect := `Trace
 run
  UseSubA`
	stackData := BriefStack(stack.Data())
	if stackData != expect {
		t.Fatalf("expect %s, but got %s", expect, stackData)
		stackJSON, _ := stack.JSON()
		t.Logf("stack: %s", stackJSON)
	}
}

func run() {
	pkg.UseSubA()
}

func BriefStack(stk *stack_model.Stack) string {
	var list []string
	for _, child := range stk.Children {
		list = append(list, briefStack(0, child))
	}
	return strings.Join(list, "\n")
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
