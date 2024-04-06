package patch

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

var a int = 123

func TestPatchVarTest(t *testing.T) {
	mock.Patch(&a, func() int {
		return 456
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched varaibel a to be %d, actual: %d", 456, b)
	}
}

func TestPatchVarWrongTypeShouldFailTest(t *testing.T) {
	var pe interface{}
	func() {
		defer func() {
			pe = recover()
		}()
		mock.Patch(&a, func() *int {
			v := 456
			return &v
		})
		b := a
		if b != 456 {
			t.Fatalf("expect patched varaibel a to be %d, actual: %d", 456, b)
		}
	}()
	expectMsg := "replacer should have type: func() int, actual: func() *int"

	if pe == nil {
		t.Fatalf("expect panic: %q, actual nil", expectMsg)
	}
	msg := fmt.Sprint(pe)
	if msg != expectMsg {
		t.Fatalf("expect err %q, actual: %q", expectMsg, msg)
	}
}

const pkgName = "github.com/xhd2015/xgo/runtime/test/patch"

func TestPatchVarByNameTest(t *testing.T) {
	mock.PatchByName(pkgName, "a", func() int {
		return 456
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
}

func TestPatchVarByNamePtrTest(t *testing.T) {
	mock.PatchByName(pkgName, "a", func() *int {
		x := 456
		return &x
	})
	pb := &a
	b := *pb
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
}

const testVersion = "1.0"

func TestPatchConstByNamePtrTest(t *testing.T) {
	mock.PatchByName(pkgName, "testVersion", func() string {
		return "1.5"
	})
	version := testVersion
	if version != "1.5" {
		t.Fatalf("expect patched version a to be %s, actual: %s", "1.5", version)
	}
}

func TestPatchConstByNameWrongTypeShouldFail(t *testing.T) {
	var pe interface{}
	func() {
		defer func() {
			pe = recover()
		}()
		mock.PatchByName(pkgName, "a", func() string {
			return "1.5"
		})
	}()
	expectMsg := "replacer should have type: func() int, actual: func() string"
	if pe == nil {
		t.Fatalf("expect panic: %q, actual nil", expectMsg)
	}
	msg := fmt.Sprint(pe)
	if msg != expectMsg {
		t.Fatalf("expect err %q, actual: %q", expectMsg, msg)
	}
}
