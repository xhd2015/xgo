package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trap"
)

// MarshalAnyJSON marshals aribitray go value `v`
// to JSON, when it encounters unmarshalable values
// like func, chan, it will bypass these values.
func MarshalAnyJSON(v interface{}) ([]byte, error) {
	funcInfo := functab.Info("encoding/json", "newTypeEncoder")
	if funcInfo == nil {
		fmt.Fprintf(os.Stderr, "WARNING: encoding/json.newTypeEncoder not trapped(requires xgo).\n")
		return json.Marshal(v)
	}

	// can be done via filter
	if false {
		v = Decyclic(v)
	}

	// get the unmarshalable function
	unmarshalable := getMarshaler(funcInfo.Func, reflect.TypeOf(unmarshalableFunc))
	var data []byte
	var err error
	// mock the encoding json
	trap.WithFuncOverride(funcInfo, &trap.Interceptor{
		Post: func(ctx context.Context, f *core.FuncInfo, args, result core.Object, data interface{}) error {
			if f != funcInfo {
				return nil
			}
			resField := result.GetFieldIndex(0)

			// if unmarshalable, replace with an empty struct
			if reflect.ValueOf(resField.Value()).Pointer() == reflect.ValueOf(unmarshalable).Pointer() {
				resField.Set(getMarshaler(funcInfo.Func, reflect.TypeOf(struct{}{})))
			}
			return nil
		},
	}, func() {
		data, err = json.Marshal(v)
	})

	return data, err
}

func unmarshalableFunc() {}

// newTypeEncoder signature: func(x Type,allowAddr bool) func()
func getMarshaler(newTypeEncoder interface{}, v reflect.Type) interface{} {
	var res interface{}
	trap.Direct(func() {
		results := reflect.ValueOf(newTypeEncoder).Call([]reflect.Value{
			reflect.ValueOf(v),
			reflect.ValueOf(false),
		})
		res = results[0].Interface()
	})
	return res
}

type decyclicer struct {
	seen map[uintptr]struct{}
}

func Decyclic(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	d := &decyclicer{
		seen: map[uintptr]struct{}{},
	}
	d.clear(reflect.ValueOf(v), func(r reflect.Value) {
		v = r.Interface()
	})
	return v
}

func makeAddrable(v reflect.Value, set func(r reflect.Value)) reflect.Value {
	if v.CanAddr() {
		return v
	}
	p := reflect.New(v.Type())
	p.Elem().Set(v)
	x := p.Elem()
	set(x)
	return x
}

func (c *decyclicer) clear(v reflect.Value, set func(r reflect.Value)) {
	// fmt.Printf("clear: %v\n", v.Type())
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return
		}

		// only pointer can create cyclic
		ptr := v.Pointer()
		if ptr == 0 {
			return
		}
		if _, ok := c.seen[ptr]; ok {
			// fmt.Printf("found : 0x%x -> %v\n", ptr, v.Interface())
			set(reflect.Zero(v.Type()))
			return
		}
		c.seen[ptr] = struct{}{}
		defer delete(c.seen, ptr)

		v = makeAddrable(v, set)
		c.clear(v.Elem(), func(r reflect.Value) {
			v.Elem().Set(r)
		})
	case reflect.Interface:
		if v.IsNil() {
			return
		}
		v = makeAddrable(v, set)
		c.clear(v.Elem(), func(r reflect.Value) {
			// NOTE: interface{} is special
			// we can directly can call v.Set
			// instead of v.Elem().Set()
			v.Set(r)
			if v.Elem().Kind() == reflect.Ptr && v.Elem().IsNil() {
				// fmt.Printf("found isNil\n")
				// avoid {nil-value,non-nil type}
				set(reflect.Zero(v.Type()))
			}
		})
	case reflect.Array, reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint8,
			reflect.Float64, reflect.Float32,
			reflect.String,
			reflect.Bool:
			return
		}
		v = makeAddrable(v, set)
		for i := 0; i < v.Len(); i++ {
			e := v.Index(i)
			c.clear(e, func(r reflect.Value) {
				e.Set(r)
			})
		}
	case reflect.Map:
		v = makeAddrable(v, set)
		iter := v.MapRange()
		// sets := [][2]reflect.Value{}
		for iter.Next() {
			vi := v.MapIndex(iter.Key())
			c.clear(vi, func(r reflect.Value) {
				v.SetMapIndex(iter.Key(), r)
			})
		}
	case reflect.Struct:
		// fmt.Printf("struct \n")
		// make struct addrable
		v = makeAddrable(v, set)

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanSet() {
				c.clear(field, func(r reflect.Value) {
					field.Set(r)
				})
			} else {
				e := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
				c.clear(e.Elem(), func(r reflect.Value) {
					e.Elem().Set(r)
				})
				// panic(fmt.Errorf("cannot set: %v", field))
			}
		}
	case reflect.Chan, reflect.Func:
		// ignore
	default:
		// int
	}
}
