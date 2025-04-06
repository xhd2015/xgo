package trace

import (
	"runtime"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/debug/util"
	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

func TestMockedFuncShouldShowInTrace(t *testing.T) {
	// testing.tRunner
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	pc := pcs[0]
	funcInfo := runtime.FuncForPC(pc)

	t.Logf("funcInfo: %s", funcInfo.Name())

	_, file, line, _ := runtime.Caller(1)
	t.Logf("called from %s:%d", file, line)
	var data []byte
	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			data, _ = stack.JSON()
		},
	}, nil, func() (interface{}, error) {
		mock.Patch(A, func() string {
			return "mock_A"
		})
		h()
		return nil, nil
	})

	// t.Logf("trace: %s", data)

	expectTraceSequence := []string{
		"{",
		`"Name":"h",`,
		`"Name":"A",`,
		`mock_A`,
		`"Name":"B",`,
		`"Name":"C",`,
		`"Name":"C",`,
		"}",
	}
	err := util.CheckSequence(string(data), expectTraceSequence)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func h() {
	A()
	B()
	C()
}

func A() string {
	return "A"
}

func B() string {
	C()
	return "B"
}

func C() string {
	return "C"
}
