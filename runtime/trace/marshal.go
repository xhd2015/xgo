package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trap"
)

// MarshalAnyJSON marshals aribitray go value `v`
// to JSON, when it encounters unmarshalable values
// like func, chan, it will bypass these values.
func MarshalAnyJSON(v interface{}) ([]byte, error) {
	newTypeEncoder := functab.Info("encoding/json", "newTypeEncoder")
	if newTypeEncoder == nil {
		fmt.Fprintf(os.Stderr, "WARNING: encoding/json.newTypeEncoder not trapped(requires xgo).\n")
		return json.Marshal(v)
	}

	// get the unmarshalable function
	unmarshalable := getMarshaler(newTypeEncoder.Func, reflect.TypeOf(unmarshalableFunc))
	var data []byte
	var err error
	// mock the encoding json
	trap.WithFuncOverride(newTypeEncoder, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (interface{}, error) {
			return nil, nil
		},
		Post: func(ctx context.Context, f *core.FuncInfo, args, result core.Object, data interface{}) error {
			if f != newTypeEncoder {
				return nil
			}
			resField := result.GetFieldIndex(0)

			// if unmarshalable, replace with an empty struct
			if reflect.ValueOf(resField.Value()).Pointer() == reflect.ValueOf(unmarshalable).Pointer() {
				resField.Set(getMarshaler(newTypeEncoder.Func, reflect.TypeOf(struct{}{})))
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
