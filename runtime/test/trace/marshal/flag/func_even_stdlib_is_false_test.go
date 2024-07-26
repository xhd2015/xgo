package marshal

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
)

// flag: --trap-stdlib=false
func TestMarshalFuncEvenStdlibIsFalse(t *testing.T) {
	data, err := trace.MarshalAnyJSON(map[string]interface{}{
		"test": test,
	})
	if err != nil {
		t.Error(err)
		return
	}
	expect := `{"test":{}}`
	if string(data) != expect {
		t.Errorf("expect: %s, actual: %s", expect, string(data))
	}
}

func test() {

}
