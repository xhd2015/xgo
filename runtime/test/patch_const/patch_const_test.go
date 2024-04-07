//go:build go1.20
// +build go1.20

package patch_const

import (
	"fmt"
	"os"
	"testing"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch_const/sub"
)

const pkgPath = "github.com/xhd2015/xgo/runtime/test/patch_const"
const subPkgPath = "github.com/xhd2015/xgo/runtime/test/patch_const/sub"
const testVersion = "1.0"

const N = 50

func TestPatchInElseShouldWork(t *testing.T) {
	if os.Getenv("nothing") == "nothing" {
		t.Fatalf("should go else")
	} else {
		mock.PatchByName(pkgPath, "N", func() int {
			return 5
		})
		b := N

		if b != 5 {
			t.Fatalf("expect b to be %d,actual: %d", 5, b)
		}
	}
}

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
		mock.PatchByName(pkgPath, "N", func() string {
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

func TestPatchConstOperationShouldCompileAndSkipMock(t *testing.T) {
	// should have effect
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	// because N is used inside an operation
	// it's type is not yet determined, so
	// should not rewrite it
	size := N * unsafe.Sizeof(int(0))
	if size != 80 {
		t.Fatalf("expect N not patched and size to be %d, actual: %d\n", 80, size)
	}
}

func TestPatchOtherPkgConstOperationShouldWork(t *testing.T) {
	// should have effect
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	// because N is used inside an operation
	// it's type is not yet determined, so
	// should not rewrite it
	size := sub.N * unsafe.Sizeof(int(0))
	if size != 80 {
		t.Fatalf("expect N not patched and size to be %d, actual: %d\n", 80, size)
	}
}

func TestConstOperationNaked(t *testing.T) {
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	var size int64 = N + 1
	if size != 11 {
		t.Fatalf("expect N not patched and size to be %d, actual: %d\n", 11, size)
	}
}

const M = 10

func TestTwoConstAdd(t *testing.T) {
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	var size int64 = (N + M) * 2
	if size != 40 {
		t.Fatalf("expect N not patched and size to be %d, actual: %d\n", 40, size)
	}
}

func TestConstOperationParen(t *testing.T) {
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	var size int64 = (N + 1) * 2
	if size != 22 {
		t.Fatalf("expect N not patched and size to be %d, actual: %d\n", 22, size)
	}
}

// local const
func TestPatchConstInAssignmentShouldWork(t *testing.T) {
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	var a int64 = N

	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchConstInAssignmentNoDefShouldWork(t *testing.T) {
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	var a int64 = 100
	if os.Getenv("nothing") == "" {
		a = N
	}

	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchConstInFuncArgShouldSkip(t *testing.T) {
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	a := f(N)
	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchConstInTypeConvertArgShouldWork(t *testing.T) {
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	a := int64(N)
	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}
func f(a int64) int64 {
	return a
}

func TestCaseConstShouldSkip(t *testing.T) {
	n := int64(50)
	switch n {
	case N:
	case N + 1:
		t.Fatalf("should not fail to N+1")
	default:
		t.Fatalf("should not fall to default")
	}

	switch n {
	case N, N + 1:
	default:
		t.Fatalf("should not fall to default")
	}
}

func TestReturnConstShouldWork(t *testing.T) {
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	n := getN()
	if n != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 50, n)
	}
}

func getN() int64 {
	return N
}

// other package
func TestPatchOtherPackageConstInAssignmentShouldWork(t *testing.T) {
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	var a int64 = sub.N

	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchOtherPackageConstInAssignmentNoDefShouldWork(t *testing.T) {
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	var a int64 = 100
	if os.Getenv("nothing") == "" {
		a = sub.N
	}

	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchOtherPackageConstInFuncArgShouldWork(t *testing.T) {
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	a := f(sub.N)
	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchOtherPackageConstInTypeConvertArgShouldWork(t *testing.T) {
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	a := int64(sub.N)
	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}
