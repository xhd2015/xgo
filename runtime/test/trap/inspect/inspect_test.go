package inspect

import (
	"testing"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/trap"
)

func TestInspectFunc(t *testing.T) {
	// plain function
	recv, f, _, _ := trap.InspectPC(F)
	if f == nil {
		t.Fatalf("F not found")
	}
	if recv != nil {
		t.Fatalf("expect receive for F is nil, actual: %v", recv)
	}

	// struct method
	s := &struct_{}
	srecv, sf, _, _ := trap.InspectPC(s.F)
	if sf == nil {
		t.Fatalf("struct_.F not found")
	}
	if srecv == nil {
		t.Fatalf("expect receive for struct_.F not nil, actual: %v", srecv)
	}
	inspectS, ok := srecv.(**struct_)
	if !ok {
		t.Fatalf("expect receive for struct_.F to be **struct_, actual: %T", srecv)
	}
	if *inspectS != s {
		t.Fatalf("expect receive for struct_.F to be 0x%x, actual: 0x%x", unsafe.Pointer(s), unsafe.Pointer(*inspectS))
	}

	// interface method
	var intf interface_ = s
	irecv, if_, _, _ := trap.InspectPC(intf.F)
	if if_ == nil {
		t.Fatalf("interface_.F not found")
	}
	// should resolving into the underlying func info
	if if_.Interface {
		t.Fatalf("expect func info of interface_.F Interface to be false, actual false")
	}
	if if_ != sf {
		t.Fatalf("expect func info of interface_.F resolve into struct_.F, actual: %v", if_)
	}
	if irecv == nil {
		t.Fatalf("expect receive for interface_.F not nil, actual: %v", irecv)
	}

	inspectIrecv, ok := irecv.(**struct_)
	if !ok {
		t.Fatalf("expect receive for interface_.F to be **struct_, actual: %T", irecv)
	}
	if *inspectIrecv != s {
		t.Fatalf("expect receive for interface_.F to be 0x%x, actual: 0x%x", unsafe.Pointer(s), unsafe.Pointer(*inspectIrecv))
	}
}

func F() {
	panic("should never call me")
}

type struct_ struct {
}

func (c *struct_) F() {
	panic("should never call me")
}

type interface_ interface {
	F()
}
