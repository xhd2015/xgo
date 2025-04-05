package trace_marshal

import (
	"reflect"
	"unsafe"
)

// cyclic are often caused by tree like data structures
//
// NOTE: this file is a temporiray backup
// decyclic seems has signaficant memory consumption
// making it slow to decyclic when encounters large data
//
// Problems with modifying decyclic:
//
//	may cause data race because value is being
//	used in another goroutine
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
	arr := []interface{}{v}
	d.clear(reflect.ValueOf(arr), func(r reflect.Value) {
		panic("should not call slice's set")
	})
	return arr[0]
}

// func makeAddrable(v reflect.Value, set func(r reflect.Value)) reflect.Value {
// 	if v.CanAddr() {
// 		return v
// 	}
// 	if false {
// 		p := reflect.New(v.Type())
// 		p.Elem().Set(v)
// 		x := p.Elem()
// 		set(x)
// 		return x
// 	}
// 	panic("not addressable")
// }

func (c *decyclicer) clear(v reflect.Value, set func(r reflect.Value)) {
	// fmt.Printf("clear: %v\n", v.Type())
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		// if !v.CanAddr() {
		// 	return
		// }

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

		c.clear(v.Elem(), func(r reflect.Value) {
			v.Elem().Set(r)
		})
	case reflect.Interface:
		if v.IsNil() {
			return
		}
		// if !v.CanAddr() {
		// 	return
		// }
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
		case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16,
			reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16,
			reflect.Float64, reflect.Float32,
			reflect.String,
			reflect.Bool:
			return
		case reflect.Int8, reflect.Uint8:
			// []byte -> Uint8
			// ignore some: 10K JSON
			n := v.Len()
			if v.Kind() == reflect.Slice && n > 10*1024 {
				// reserve first 16 and last 16
				//   S[:16] ... + S[len(S)-16:]
				const reserve = 16
				const ellipse = 3
				const totalLen = reserve*2 + ellipse
				newSlice := reflect.MakeSlice(v.Type(), totalLen, totalLen)
				for i := 0; i < reserve; i++ {
					newSlice.Index(i).Set(v.Index(i))
				}
				for i := 0; i < ellipse; i++ {
					if v.Kind() == reflect.Uint8 {
						newSlice.Index(reserve + i).SetUint('.')
					} else if v.Kind() == reflect.Int8 {
						newSlice.Index(reserve + i).SetInt('.')
					}
				}
				for i := 0; i < reserve; i++ {
					newSlice.Index(reserve + ellipse + i).Set(v.Index(n - reserve + i))
				}
				set(newSlice)
			}
			return
		}
		for i := 0; i < v.Len(); i++ {
			e := v.Index(i)
			c.clear(e, func(r reflect.Value) {
				e.Set(r)
			})
		}
	case reflect.Map:
		if !v.CanAddr() {
			return
		}
		iter := v.MapRange()
		// sets := [][2]reflect.Value{}
		for iter.Next() {
			vi := v.MapIndex(iter.Key())
			c.clear(vi, func(r reflect.Value) {
				v.SetMapIndex(iter.Key(), r)
			})
		}
	case reflect.Struct:
		if !v.CanAddr() {
			return
		}
		// fmt.Printf("struct \n")
		// make struct addrable
		// v = makeAddrable(v, set)
		// fmt.Printf("struct can addr: %v\n", v.CanAddr())
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			// fmt.Printf("field: %v %v\n", field, field.CanAddr())
			if field.CanSet() && false {
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
