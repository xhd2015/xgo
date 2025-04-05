package trace_marshal

import (
	"fmt"
	"reflect"
	"testing"
)

type _struct struct {
	self *_struct
}

func TestDecyclicUnexported(t *testing.T) {
	s := &_struct{}
	s.self = s

	vs := Decyclic(s)
	if vs.(*_struct).self != nil {
		t.Fatalf("expect vs.self to be nil")
	}
}

type _structMap struct {
	m map[string]*_structMap
}

func TestDecyclicMapValue(t *testing.T) {
	s := &_structMap{
		m: make(map[string]*_structMap),
	}
	s.m["self"] = s

	v1 := reflect.ValueOf(s).Interface()
	if v1.(*_structMap).m["self"] == nil {
		t.Fatalf("expect v1.m['self'] to be non nil")
	}

	vs := Decyclic(s)
	if vs.(*_structMap).m["self"] != nil {
		t.Fatalf("expect vs.m['self'] to be nil")
	}
}

type _structWithIntf struct {
	self interface{}
}

func TestDecyclicInterface(t *testing.T) {
	s := &_structWithIntf{}
	s.self = s

	vs := Decyclic(s)
	if vs.(*_structWithIntf).self != nil {
		t.Fatalf("expect vs.self to be nil")
	}
}

type _valueStruct struct {
	A int
}

func TestDecyclicValueStruct(t *testing.T) {
	s := _valueStruct{
		A: 123,
	}

	vs := Decyclic(s)
	vsA := vs.(_valueStruct).A
	if vsA != 123 {
		t.Fatalf("vs.A: %v", vsA)
	}
}

func TestDecyclicMapInterface(t *testing.T) {
	m := make(map[string]interface{})
	m["A"] = 123

	vs := Decyclic(m)
	vsA := vs.(map[string]interface{})["A"]
	if fmt.Sprint(vsA) != "123" {
		t.Fatalf("vs.A: %v", vsA)
	}
}

type byteSliceHolder struct {
	slice []byte
}

func TestDecyclicLargeByteSlice(t *testing.T) {
	s := &byteSliceHolder{
		slice: make([]byte, 10*1024+1),
	}
	vs := Decyclic(s)
	vslice := vs.(*byteSliceHolder).slice

	// 16 + ... + 16
	if len(vslice) != 35 {
		t.Logf("expect len to be %d, actual: %v", 35, len(vslice))
	}
}
