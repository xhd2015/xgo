package cyclic

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

type Ref struct {
	Name     string
	Children []*Ref
}

func makeCyclicRef() *Ref {
	ref := &Ref{
		Name: "root",
		Children: []*Ref{
			{
				Name: "child",
			},
		},
	}

	ref.Children[0].Children = []*Ref{ref}
	return ref
}

const supportCyclic = false

func TestMarshalCyclic(t *testing.T) {
	var exportedStack stack_model.IStack
	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			exportedStack = stack
		},
	}, nil, func() (interface{}, error) {
		return makeCyclicRef(), nil
	})

	data, err := exportedStack.JSON()
	if err != nil {
		t.Fatal(err)
	}
	dataStr := string(data)
	if supportCyclic {
		n := strings.Count(dataStr, `"Name":"child"`)
		if n != 2 {
			t.Errorf("expect 2 child, but got %d", n)
		}
		n = strings.Count(dataStr, `"Name":"root"`)
		if n != 1 {
			t.Errorf("expect 1 root, but got %d", n)
		}
	} else {
		expectContains := `unsupported value: encountered a cycle via []*cyclic.Ref`
		if !strings.Contains(dataStr, expectContains) {
			t.Errorf("expect contains %q, but not found", expectContains)
		}
	}
}
