package record

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/support/assert"
)

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

func TestRecord(t *testing.T) {
	var records []string
	trace.RecordCall(A, func(res *string) {
		records = append(records, fmt.Sprintf("A is called: %s", *res))
	})
	trace.RecordResult(B, func(res *string) {
		records = append(records, fmt.Sprintf("B is called: %s", *res))
	})
	trace.RecordCall(C, func(res *string) {
		records = append(records, fmt.Sprintf("C is called: %s", *res))
	})
	h()
	expected := "A is called: \n" +
		"C is called: \n" +
		"B is called: B\n" +
		"C is called: "
	if diff := assert.Diff(expected, strings.Join(records, "\n")); diff != "" {
		t.Error(diff)
	}
}
