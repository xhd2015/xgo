// Package loose pins the modern loose-JSON contract used by xgo trace:
// unsupported types (func, chan, …) must marshal as placeholders under
// runtime.MarshalNoError, not soft-fail into {"error":"…unsupported…"}.
//
// Expected behavior with an instrumented GOROOT:
//   - go1.26 and earlier (classic encoding/json + unsupportedTypeEncoder patch): PASS
//   - go1.27 default (json v2 / makeInvalidArshaler, patch not yet applied): FAIL
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
	// (go1.27 default today: json v2 path, no loose hook on makeInvalidArshaler).
	if strings.Contains(data, `"error"`) && strings.Contains(strings.ToLower(data), "unsupported") {
		t.Fatalf("loose JSON marshaling did not engage: got soft-fail error form %q\n"+
			"hint: on go1.27+ default (json v2), encoding/json/encode.go patch is inactive; "+
			"need a makeInvalidArshaler (or equivalent) hook", data)
	}

	// Path-A placeholder from patched unsupportedTypeEncoder:
	//   e.WriteString(fmt.Sprintf("{%q:%q}", v.Type().String(), "?"))
	// e.g. {"func()":"?"} or {"chan int":"?"}
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
