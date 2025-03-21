package trace_runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

func tryRemoveFirstCtx(args []runtime.XgoField) []runtime.XgoField {
	if len(args) == 0 {
		return args
	}
	if _, ok := args[0].Ptr.(*context.Context); ok {
		return args[1:]
	}
	return args
}

func trySplitLastError(results []runtime.XgoField) ([]runtime.XgoField, error) {
	n := len(results)
	if n == 0 {
		return results, nil
	}
	res := results[n-1]
	if res.Ptr == nil {
		return results, nil
	}

	if ptrErr, ok := res.Ptr.(*error); ok {
		if ptrErr == nil {
			return results[:n-1], nil
		}
		return results[:n-1], *ptrErr
	}
	return results, nil
}

func marshalNoError(v interface{}) (result []byte) {
	g := GetG()
	if !g.looseJsonMarshaling {
		g.looseJsonMarshaling = true
		defer func() {
			g.looseJsonMarshaling = false
		}()
	}
	var err error
	defer func() {
		var stackTrace []byte
		if e := recover(); e != nil {
			stackTrace = debug.Stack()
			if pe, ok := e.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
		var qstackTrace string
		if len(stackTrace) > 0 {
			qstackTrace = fmt.Sprintf(", %q: %q", "stackTrace", string(stackTrace))
		}
		if err != nil {
			result = []byte(fmt.Sprintf("{%q: %q%s}", "error", err.Error(), qstackTrace))
		}
	}()
	result, err = json.Marshal(v)
	return
}

type StructValue []runtime.XgoField

func (c StructValue) MarshalJSON() ([]byte, error) {
	fields := make([]string, len(c))
	for i, v := range c {
		val, err := json.Marshal(v.Ptr)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal field %s: %w", v.Name, err)
		}
		fields[i] = fmt.Sprintf("%q: %s", v.Name, val)
	}
	return []byte(fmt.Sprintf("{%s}", strings.Join(fields, ","))), nil
}
