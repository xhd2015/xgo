//go:build go1.20
// +build go1.20

package patch_const

import (
	"fmt"
	"os"
	"testing"
	"unsafe"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_const/sub"
)

const pkgPath = "github.com/xhd2015/xgo/runtime/test/patch_const"
const subPkgPath = "github.com/xhd2015/xgo/runtime/test/patch_const/sub"
const testVersion = "1.0"

const N = 50

func TestPatchInElseShouldWork(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	mock.PatchByName(pkgPath, "testVersion", func() string {
		return "1.5"
	})
	version := testVersion
	if version != "1.5" {
		t.Fatalf("expect patched version a to be %s, actual: %s", "1.5", version)
	}
}

func TestPatchConstByNameWrongTypeShouldError(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	var size int64 = (N + M) * 2
	if size != 40 {
		t.Fatalf("expect N not patched and size to be %d, actual: %d\n", 40, size)
	}
}

func TestConstOperationParen(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	var a int64 = N

	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchConstInAssignmentNoDefShouldWork(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	mock.PatchByName(pkgPath, "N", func() int {
		return 10
	})
	a := f(N)
	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchConstInTypeConvertArgShouldWork(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	n := int64(50)
	switch n {
	case N:
	case N + 1:
		t.Fatalf("should not error to N+1")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	var a int64 = sub.N

	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchOtherPackageConstInAssignmentNoDefShouldWork(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
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
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	a := f(sub.N)
	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

func TestPatchOtherPackageConstInTypeConvertArgShouldWork(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	mock.PatchByName(subPkgPath, "N", func() int {
		return 10
	})
	a := int64(sub.N)
	if a != 10 {
		t.Fatalf("expect a to be %d, actual: %d\n", 10, a)
	}
}

const x = "123"

func TestPatchConstOverlappingNameShouldSkip(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	x := make([]int, 0, 10)
	x = append(x, 10)
	if x[0] != 10 {
		t.Fatalf("expect x[0] to be %d, actual: %d", 10, x[0])
	}
}

func exampleSprintf(args ...interface{}) string {
	return fmt.Sprintf("%v", args...)
}
func TestPatchLitPlusLitShouldCompile(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	s := "should "
	a := exampleSprintf(s + "compile")
	b := exampleSprintf("should " + "compile") // this skips from wrapping
	if a != b {
		t.Fatalf("expect exampleSprintf result to be %q, actual: %q", b, a)
	}
}

type Label string

func convertLabel(label Label) string {
	return string(label)
}
func TestPatchOtherPkgConst(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	label := convertLabel(sub.LabelPrefix + "v2")
	if label != "label:v2" {
		t.Fatalf("bad label: %s", label)
	}
}

const good = 2
const reason = "test"

func TestNameConflictWithArgShouldSkip(t *testing.T) {
	t.Skip("constant patching has been prohibited since xgo v1.1.0")
	reasons := getReasons("good")
	if len(reasons) != 2 || reasons[0] != "ok" || reasons[1] != "good" {
		t.Fatalf("bad reason: %v", reasons)
	}

	getReasons2 := func(good string) (reason []string) {
		reason = append(reason, "ok")
		reason = append(reason, good)
		return
	}
	reasons2 := getReasons2("good")
	if len(reasons2) != 2 || reasons2[0] != "ok" || reasons2[1] != "good" {
		t.Fatalf("bad reason2: %v", reasons2)
	}
}

func getReasons(good string) (reason []string) {
	reason = append(reason, "ok")
	reason = append(reason, good)
	return
}
