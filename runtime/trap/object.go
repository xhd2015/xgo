package trap

import (
	"fmt"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core"
)

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
