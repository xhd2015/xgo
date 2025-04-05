package patch

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/internal/legacy"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/sub"
)

func TestPatchVarTest(t *testing.T) {
	mock.Patch(&a, func() int {
		return 456
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
}

func TestPatchVarPtrTest(t *testing.T) {
	mock.Patch(&a, func() *int {
		v := 333
		return &v
	})
	b := &a
	if *b != 333 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 333, *b)
	}
}

// TestPatchVarAndPtrTestSameVariable also validates
// goroutine-separation: it does not interfere with TestPatchVarPtrTest
func TestPatchVarAndPtrTestSameVariable(t *testing.T) {
	mock.Patch(&a, func() int {
		return 456
	})
	mock.Patch(&a, func() *int {
		v := 789
		return &v
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
	c := &a
	if *c != 789 {
		t.Fatalf("expect patched variable ptr a to be %d, actual: %d", 789, *c)
	}

	// read again
	if a != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, a)
	}
	if *c != 789 {
		t.Fatalf("expect patched variable ptr a to be %d, actual: %d", 789, *c)
	}
}
func TestPatchVarAndPtrTestNewVariable(t *testing.T) {
	mock.Patch(&a2, func() int {
		return 456
	})
	mock.Patch(&a2, func() *int {
		v := 789
		return &v
	})
	b := a2
	if b != 456 {
		t.Fatalf("expect patched variable a2 to be %d, actual: %d", 456, b)
	}
	c := &a2
	if *c != 789 {
		t.Fatalf("expect patched variable ptr a2 to be %d, actual: %d", 789, *c)
	}

	// read again
	if a2 != 456 {
		t.Fatalf("expect patched variable a2 to be %d, actual: %d", 456, a2)
	}
	if *c != 789 {
		t.Fatalf("expect patched variable ptr a2 to be %d, actual: %d", 789, *c)
	}
}

func TestPatchVarPtrFallbackTest(t *testing.T) {
	mock.Patch(&a, func() int {
		return 456
	})
	b := &a
	if *b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, *b)
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
			t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
		}
	}()
	if legacy.V1_0_0 {
		expectMsg := "replacer should have type: func() int, actual: func() *int"
		if pe == nil {
			t.Fatalf("expect panic: %q, actual nil", expectMsg)
		}
		msg := fmt.Sprint(pe)
		if msg != expectMsg {
			t.Fatalf("expect err %q, actual: %q", expectMsg, msg)
		}
	}
}

const pkgPath = "github.com/xhd2015/xgo/runtime/test/patch"
const subPkgPath = "github.com/xhd2015/xgo/runtime/test/patch/sub"

func TestPatchVarByNameTest(t *testing.T) {
	if !legacy.V1_0_0 {
		t.Skip("PatchByName is deprecated and no longer supported, use Patch instead")
	}
	mock.PatchByName(pkgPath, "a", func() int {
		return 456
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
}

func TestPatchVarByNamePtrTest(t *testing.T) {
	if !legacy.V1_0_0 {
		t.Skip("PatchByName is deprecated and no longer supported, use Patch instead")
	}
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

func TestPatchSwitchCaseShouldCompile(t *testing.T) {
	toJSONRaw(10)
}

func toJSONRaw(v interface{}) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	switch v := v.(type) {
	case []byte:
		return v, nil
	case json.RawMessage:
		return v, nil
	case string:
		return json.RawMessage([]byte(v)), nil
	default:
		return json.Marshal(v)
	}
}

func TestMakeInOtherPackageShouldCompile(t *testing.T) {
	// previous error:sub.NameSet (type) is not an expression
	set := make(sub.NameSet)
	_ = set
}
