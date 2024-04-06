package patch

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/sub"
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

const pkgPath = "github.com/xhd2015/xgo/runtime/test/patch"
const subPkgPath = "github.com/xhd2015/xgo/runtime/test/patch/sub"

func TestPatchVarByNameTest(t *testing.T) {
	mock.PatchByName(pkgPath, "a", func() int {
		return 456
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
}

func TestPatchVarByNamePtrTest(t *testing.T) {
	mock.PatchByName(pkgPath, "a", func() *int {
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
	mock.PatchByName(pkgPath, "testVersion", func() string {
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
		mock.PatchByName(pkgPath, "a", func() string {
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

const N = 50

func TestPatchConstOperationShouldCompileAndSkipMock(t *testing.T) {
	// should have no effect
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	// because N is used inside an operation
	// it's type is not yet determined, so
	// should not rewrite it
	size := N * unsafe.Sizeof(int(0))
	if size != 400 {
		t.Logf("expect N not patched and size to be %d, actual: %d\n", 400, size)
	}
}

func TestPatchOtherPkgConstOperationShouldCompileAndSkipMock(t *testing.T) {
	// should have no effect
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	// because N is used inside an operation
	// it's type is not yet determined, so
	// should not rewrite it
	size := sub.N * unsafe.Sizeof(int(0))
	if size != 400 {
		t.Logf("expect N not patched and size to be %d, actual: %d\n", 400, size)
	}
}
