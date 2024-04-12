package trace_marshal

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
)

type _struct struct {
	self *_struct
}

func TestDecyclicUnexported(t *testing.T) {
	s := &_struct{}
	s.self = s

	vs := trace.Decyclic(s)
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

	vs := trace.Decyclic(s)
	if vs.(*_structMap).m["self"] != nil {
		t.Fatalf("expect vs.m['self'] to be nil")
	}
}

type _interface struct {
	self interface{}
}

func TestDecyclicInterface(t *testing.T) {
	s := &_interface{}
	s.self = s

	vs := trace.Decyclic(s)
	if vs.(*_interface).self != nil {
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

	vs := trace.Decyclic(s)
	vsA := vs.(_valueStruct).A
	if vsA != 123 {
		t.Fatalf("vs.A: %v", vsA)
	}
}

func TestDecyclicMapInterface(t *testing.T) {
	m := make(map[string]interface{})
	m["A"] = 123

	vs := trace.Decyclic(m)
	vsA := vs.(map[string]interface{})["A"]
	if fmt.Sprint(vsA) != "123" {
		t.Fatalf("vs.A: %v", vsA)
	}
}
