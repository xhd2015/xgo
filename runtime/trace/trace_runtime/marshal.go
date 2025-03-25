package trace_runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
)

func tryRemoveFirstCtx(argNames []string, args []interface{}) ([]string, []interface{}) {
	if len(args) == 0 {
		return argNames, args
	}
	if _, ok := args[1].(*context.Context); ok {
		return argNames[1:], args[1:]
	}
	return argNames, args
}

func trySplitLastError(resultNames []string, results []interface{}) ([]string, []interface{}, error) {
	n := len(results)
	if n == 0 {
		return resultNames, results, nil
	}
	res := results[n-1]
	if res == nil {
		return resultNames, results, nil
	}

	if ptrErr, ok := res.(*error); ok {
		if ptrErr == nil {
			return resultNames[:n-1], results[:n-1], nil
		}
		return resultNames[:n-1], results[:n-1], *ptrErr
	}
	return resultNames, results, nil
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

type structValue struct {
	Names  []string
	Values []interface{}
}

func newStructValue(names []string, values []interface{}) *structValue {
	return &structValue{
		Names:  names,
		Values: values,
	}
}

func (c *structValue) MarshalJSON() ([]byte, error) {
	fields := make([]string, len(c.Names))
	for i, name := range c.Names {
		val, err := json.Marshal(c.Values[i])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal field %s: %w", name, err)
		}
		fields[i] = fmt.Sprintf("%q: %s", name, val)
	}
	return []byte(fmt.Sprintf("{%s}", strings.Join(fields, ","))), nil
}
