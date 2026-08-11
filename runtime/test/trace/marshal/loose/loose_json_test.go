// Package loose pins the modern loose-JSON contract used by xgo trace:
// unsupported types (func, chan, …) must marshal as placeholders under
// runtime.MarshalNoError, not soft-fail into {"error":"…unsupported…"}.
//
// Expected behavior with an instrumented GOROOT:
//   - go1.26 and earlier: classic encoding/json + unsupportedTypeEncoder Path-A patch
//   - go1.27 default (json v2): makeInvalidArshaler Path-A patch in encoding/json/v2
//
// Run via:
//
//	go run ./script/run-test --include go1.26 -run TestLooseJsonMarshal ./runtime/test/trace/marshal/loose
//	go run ./script/run-test --include go1.27rc2 -run TestLooseJsonMarshal ./runtime/test/trace/marshal/loose
package loose

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
	"github.com/xhd2015/xgo/runtime/trace/stack_model"
)

// holder mixes a marshalable field with an unsupported one so nested
// encoding is exercised (not only a bare unsupported root value).
type holder struct {
	N int
	F func()
	C chan int
}

func TestLooseJsonMarshalUnsupportedFunc(t *testing.T) {
	assertLoosePlaceholder(t, func() (interface{}, error) {
		return func() {}, nil
	}, "func()")
}

func TestLooseJsonMarshalUnsupportedChan(t *testing.T) {
	assertLoosePlaceholder(t, func() (interface{}, error) {
		return make(chan int), nil
	}, "chan int")
}

func TestLooseJsonMarshalUnsupportedNested(t *testing.T) {
	assertLoosePlaceholder(t, func() (interface{}, error) {
		return holder{N: 42, F: func() {}, C: make(chan int)}, nil
	}, "func()", "chan int")
	// Nested case: the int field must survive when loose marshaling works.
	data := captureTraceJSON(t, func() (interface{}, error) {
		return holder{N: 42, F: func() {}, C: make(chan int)}, nil
	})
	if !strings.Contains(data, `"N":42`) && !strings.Contains(data, `"N": 42`) {
		// Soft-fail of the whole value often loses the sibling field.
		t.Errorf("expect nested int field N=42 preserved under loose marshaling, got %q", data)
	}
}

func assertLoosePlaceholder(t *testing.T, body func() (interface{}, error), typeHints ...string) {
	t.Helper()
	data := captureTraceJSON(t, body)

	// MarshalNoError soft-fail form when the stdlib encoder rejects the value
	// (regression: missing Path-A hook on classic unsupportedTypeEncoder or
	// go1.27+ makeInvalidArshaler under json v2).
	if strings.Contains(data, `"error"`) && strings.Contains(strings.ToLower(data), "unsupported") {
		t.Fatalf("loose JSON marshaling did not engage: got soft-fail error form %q\n"+
			"hint: classic path uses unsupportedTypeEncoder; go1.27+ default json v2 uses "+
			"makeInvalidArshaler — both need XgoIsLooseJsonMarshaling Path-A hooks", data)
	}

	// Path-A placeholder (classic unsupportedTypeEncoder or v2 makeInvalidArshaler):
	//   {"func()":"?"} / {"chan int":"?"}
	if !strings.Contains(data, `"?`) && !strings.Contains(data, `":"?"`) && !strings.Contains(data, `": "?"`) {
		t.Fatalf("expect loose placeholder containing \"?\", got %q", data)
	}
	for _, hint := range typeHints {
		if hint == "" {
			continue
		}
		if !strings.Contains(data, hint) {
			t.Errorf("expect type hint %q in loose placeholder JSON, got %q", hint, data)
		}
	}
}

func captureTraceJSON(t *testing.T, body func() (interface{}, error)) string {
	t.Helper()
	var exported stack_model.IStack
	trace.Trace(trace.Config{
		OnFinish: func(stack stack_model.IStack) {
			exported = stack
		},
	}, nil, body)
	if exported == nil {
		t.Fatal("OnFinish did not capture stack")
	}
	data, err := exported.JSON()
	if err != nil {
		t.Fatalf("stack.JSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("stack.JSON returned empty")
	}
	return string(data)
}
