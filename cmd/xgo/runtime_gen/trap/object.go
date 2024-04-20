package trap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/xhd2015/xgo/runtime/core"
)

// TODO: implement JSON.marshal on func,chan types

type object []field

type field struct {
	name   string
	valPtr interface{}
}

type objectWithErr struct {
	object
	err field
}

var _ core.Object = (object)(nil)
var _ core.ObjectWithErr = (*objectWithErr)(nil)
var _ core.Field = field{}

func appendFields(obj object, ptrs []interface{}, names []string) object {
	for i, arg := range ptrs {
		var argName string
		if i < len(names) {
			argName = names[i]
		}
		obj = append(obj, field{
			name:   argName,
			valPtr: arg,
		})
	}
	return obj
}

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

func (c *objectWithErr) GetErr() core.Field {
	return c.err
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

func (c object) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}
	var buf bytes.Buffer
	buf.WriteRune('{')
	for i, field := range c {
		name := field.name
		if name == "" {
			name = "field_" + strconv.FormatInt(int64(i), 10)
		}
		buf.WriteString(strconv.Quote(name))
		buf.WriteRune(':')
		val, err := json.Marshal(field.valPtr)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", name, err)
		}
		buf.Write(val)
		if i < len(c)-1 {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}
