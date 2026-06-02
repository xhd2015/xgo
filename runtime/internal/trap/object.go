package trap

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

// default size to shrink 1M
const DEFAULT_SIZE_LIMIT = 1 * 1024 * 1024

type object []field

type field struct {
	name   string
	valPtr interface{}
}

var _ core.Object = (object)(nil)
var _ core.Field = field{}

func (c object) GetField(name string) core.Field {
	for _, field := range c {
		if field.name == name {
			return field
		}
	}
	panic(fmt.Errorf("no field: %s", name))
}

func (c object) GetFieldIndex(i int) core.Field {
	return c[i]
}

func (c object) NumField() int {
	return len(c)
}

func (c field) Name() string {
	return c.name
}

func (c field) Set(val interface{}) {
	// if val is nil, then reflect.ValueOf(val)
	// is invalid
	if val == nil {
		// clear
		reflect.ValueOf(c.valPtr).Elem().Set(reflect.Zero(reflect.TypeOf(c.valPtr).Elem()))
		return
	}
	reflect.ValueOf(c.valPtr).Elem().Set(reflect.ValueOf(val))
}
func (c field) Ptr() interface{} {
	return c.valPtr
}

func (c field) Value() interface{} {
	return reflect.ValueOf(c.valPtr).Elem().Interface()
}

func tryRemoveFirstCtx(argNames []string, args []interface{}) ([]string, []interface{}) {
	if len(args) == 0 {
		return argNames, args
	}
	if _, ok := args[0].(*context.Context); ok {
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
		fieldName := name
		if name == "" {
			fieldName = fmt.Sprintf("__field_%d", i)
		}
		res := runtime.MarshalNoError(c.Values[i])
		if len(res) > DEFAULT_SIZE_LIMIT {
			fields[i] = fmt.Sprintf("%q: %q", fieldName, string(res[:DEFAULT_SIZE_LIMIT]))
		} else {
			fields[i] = fmt.Sprintf("%q: %s", fieldName, res)
		}
	}
	return []byte(fmt.Sprintf("{%s}", strings.Join(fields, ","))), nil
}

type ptrType int

const (
	ptrType_Value ptrType = iota
	ptrType_Ptr
)

func (c ptrType) get(p reflect.Value) reflect.Value {
	if c == ptrType_Value {
		return p.Elem()
	}
	return p
}

func resolveArgTypes(t reflect.Type, argTypes []reflect.Type) ([]ptrType, bool) {
	// assume t is func type
	if t.NumIn() != len(argTypes) {
		return nil, false
	}
	res := make([]ptrType, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		argType := argTypes[i]
		tArg := t.In(i)
		pt := ptrType_Value
		if tArg != argType {
			if tArg.Kind() != reflect.Ptr || tArg.Elem() != argType {
				return nil, false
			}
			pt = ptrType_Ptr
		}
		res[i] = pt
	}
	return res, true
}

// a,b must be func type
func checkFuncTypeMatch(a reflect.Type, b reflect.Type, skipAFirst bool) (atype string, btype string, match bool) {
	na := a.NumIn()
	nb := b.NumIn()

	base := 0
	if skipAFirst {
		base++
	}
	if na-base != nb {
		return formatFuncType(a, skipAFirst), formatFuncType(b, false), false
	}

	for i := 0; i < na; i++ {
		ta := a.In(i + base)
		tb := b.In(i)
		if ta != tb {
			return formatFuncType(a, skipAFirst), formatFuncType(b, false), false
		}
	}

	nouta := a.NumOut()
	noutb := b.NumOut()
	if nouta != noutb {
		return formatFuncType(a, skipAFirst), formatFuncType(b, false), false
	}
	for i := 0; i < nouta; i++ {
		ta := a.Out(i)
		tb := b.Out(i)
		if ta != tb {
			return formatFuncType(a, skipAFirst), formatFuncType(b, false), false
		}
	}
	return "", "", true
}

func formatFuncType(f reflect.Type, skipFirst bool) string {
	n := f.NumIn()
	i := 0
	if skipFirst {
		i++
	}
	var strBuilder strings.Builder
	strBuilder.WriteString("func(")
	for ; i < n; i++ {
		t := f.In(i)
		strBuilder.WriteString(t.String())
		if i < n-1 {
			strBuilder.WriteString(",")
		}
	}
	strBuilder.WriteString(")")

	nout := f.NumOut()
	if nout > 0 {
		strBuilder.WriteString(" ")
		if nout > 1 {
			strBuilder.WriteString("(")
		}
		for i := 0; i < nout; i++ {
			t := f.Out(i)
			strBuilder.WriteString(t.String())
			if i < nout-1 {
				strBuilder.WriteString(",")
			}
		}
		if nout > 1 {
			strBuilder.WriteString(")")
		}
	}

	return strBuilder.String()
}
