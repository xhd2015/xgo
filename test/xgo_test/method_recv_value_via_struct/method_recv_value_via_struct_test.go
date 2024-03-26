package method_recv_value_via_struct

import (
	"testing"
	"unsafe"
)

type large [512]int

type small int

type struct_ struct {
	name  string
	value int
}

func (c large) String() string {
	return "large"
}
func (c small) String() string {
	return "small"
}
func (c *struct_) String() string {
	return "struct_"
}

func TestMethodRecvValueShouldBeCopied(t *testing.T) {
	type _largeMethodValue struct {
		pc   uintptr
		recv [512]int
	}
	type _smallMethodValue struct {
		pc   uintptr
		recv int
	}
	type _structMethodValue struct {
		pc   uintptr
		recv *struct_
	}
	var lg large
	var sm small
	var st = &struct_{
		name: "init",
	}

	lgM := lg.String
	smM := sm.String
	stM := st.String

	lgv := (*(**_largeMethodValue)(unsafe.Pointer(&lgM)))
	smv := (*(**_smallMethodValue)(unsafe.Pointer(&smM)))
	stv := (*(**_structMethodValue)(unsafe.Pointer(&stM)))

	lg[0] = 1
	lgv.recv[0] = 2
	if lg[0] == lgv.recv[0] {
		t.Fatalf("expect value receiver lg do not share")
	}

	sm = 20
	smv.recv = 30
	if sm == small(smv.recv) {
		t.Fatalf("expect value receiver sm do not share")
	}

	stv.recv.name = "init2"
	if st.name != stv.recv.name {
		t.Fatalf("expect pointer receiver st share")
	}

	if unsafe.Pointer(stv.recv) != unsafe.Pointer(st) {
		t.Fatalf("expect pointer receiver st the same")
	}
}
