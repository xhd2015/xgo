package trace_variable

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/debug/util"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

var a int

func TestTraceVariable(t *testing.T) {
	var record stack_model.IStack
	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			record = stack
		},
	}, nil, func() (interface{}, error) {
		return checkA(), nil
	})
	if record == nil {
		t.Fatal("record is nil")
	}
	stk := record.Data()
	expected := `
Trace
 checkA
  a
`
	expected = strings.TrimSpace(expected)
	actual := util.BriefStack(stk)
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
		jsonData, err := json.Marshal(stk)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("stack: %v", string(jsonData))
	}

}

func checkA() bool {
	return a == 0
}
