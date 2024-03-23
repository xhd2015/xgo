package method_fm_suffix

import (
	"reflect"
	"runtime"
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

func TestMethodShouldHaveFMSuffix(t *testing.T) {
	type _methodValue struct {
		pc uintptr
	}
	lg := (large{}).String
	sm := (small(10)).String
	st := (&(struct_{})).String

	// TODO: add a document to explain why
	// &lg is **_methodValue
	lgpc := (*(**_methodValue)(unsafe.Pointer(&lg))).pc
	smpc := (*(**_methodValue)(unsafe.Pointer(&sm))).pc
	stpc := (*(**_methodValue)(unsafe.Pointer(&st))).pc

	rvlgpc := reflect.ValueOf(lg).Pointer()
	rvsmpc := reflect.ValueOf(sm).Pointer()
	rvstpc := reflect.ValueOf(st).Pointer()

	if lgpc != rvlgpc {
		t.Fatalf("expect lgpc to be 0x%x(reflect), actual: 0x%x", rvlgpc, lgpc)
	}
	if smpc != rvsmpc {
		t.Fatalf("expect smpc to be 0x%x(reflect), actual: 0x%x", rvsmpc, smpc)
	}
	if stpc != rvstpc {
		t.Fatalf("expect stpc to be 0x%x(reflect), actual: 0x%x", rvstpc, stpc)
	}

	lgName := runtime.FuncForPC(lgpc).Name()
	smName := runtime.FuncForPC(smpc).Name()
	stName := runtime.FuncForPC(stpc).Name()

	// t.Logf("lgName: %s", lgName)
	// t.Logf("smName: %s", smName)
	// t.Logf("stName: %s", stName)

	// NOTE: method values have -fm suffix
	explgname := "github.com/xhd2015/xgo/test/xgo_test/method_fm_suffix.large.String-fm"
	expsmName := "github.com/xhd2015/xgo/test/xgo_test/method_fm_suffix.small.String-fm"
	expstName := "github.com/xhd2015/xgo/test/xgo_test/method_fm_suffix.(*struct_).String-fm"

	if lgName != explgname {
		t.Fatalf("expect lgName to be %s, actual: %s", explgname, lgName)
	}
	if smName != expsmName {
		t.Fatalf("expect smName to be %s, actual: %s", expsmName, smName)
	}
	if stName != expstName {
		t.Fatalf("expect stName to be %s, actual: %s", expstName, stName)
	}

	lgProtoName := runtime.FuncForPC(reflect.ValueOf(large.String).Pointer()).Name()
	smProtoName := runtime.FuncForPC(reflect.ValueOf(small.String).Pointer()).Name()
	stProtoName := runtime.FuncForPC(reflect.ValueOf((*struct_).String).Pointer()).Name()

	// t.Logf("lgProtoName: %s", lgProtoName)
	// t.Logf("smProtoName: %s", smProtoName)
	// t.Logf("stProtoName: %s", stProtoName)

	explgProtoName := "github.com/xhd2015/xgo/test/xgo_test/method_fm_suffix.large.String"
	expsmProtoName := "github.com/xhd2015/xgo/test/xgo_test/method_fm_suffix.small.String"
	expstProtoName := "github.com/xhd2015/xgo/test/xgo_test/method_fm_suffix.(*struct_).String"

	if lgProtoName != explgProtoName {
		t.Fatalf("expect lgProtoName to be %s, actual: %s", explgProtoName, lgProtoName)
	}
	if smProtoName != expsmProtoName {
		t.Fatalf("expect smProtoName to be %s, actual: %s", expsmProtoName, smProtoName)
	}
	if stProtoName != expstProtoName {
		t.Fatalf("expect stProtoName to be %s, actual: %s", expstProtoName, stProtoName)
	}

	if lgName != lgProtoName+"-fm" {
		t.Fatalf("expect lgName prefix-fm to be %s, actual: %s", lgProtoName+"-fm", lgName)
	}
	if smName != smProtoName+"-fm" {
		t.Fatalf("expect smName prefix-fm to be %s, actual: %s", smProtoName+"-fm", smName)
	}
	if stName != stProtoName+"-fm" {
		t.Fatalf("expect stName prefix-fm to be %s, actual: %s", stProtoName+"-fm", stName)
	}
}
