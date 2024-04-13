package trace

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/util"
	"github.com/xhd2015/xgo/runtime/trace"
)

func TestMockedFuncShouldShowInTrace(t *testing.T) {
	var root *trace.Root
	trace.Options().OnComplete(func(r *trace.Root) {
		root = r
	}).Collect(func() {
		mock.Patch(A, func() string {
			return "mock_A"
		})
		h()
	})
	data, err := trace.MarshalAnyJSON(root.Export(nil))
	if err != nil {
		t.Fatal(err)
	}

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
	err = util.CheckSequence(string(data), expectTraceSequence)
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
