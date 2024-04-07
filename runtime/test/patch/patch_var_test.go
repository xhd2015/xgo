package patch

import (
	"encoding/json"
	"fmt"
	"testing"

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
