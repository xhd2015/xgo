package marshal

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
)

// go run -tags dev ./cmd/xgo test --project-dir runtime/test/trace/marshal/exclue --mock-rule='{"pkg":"encoding/json","name":"newTypeEncoder","action":"exclude"}' -run TestMarshalFuncEvenActionExclude

// flag: --mock-rule='{"pkg":"encoding/json","name":"newTypeEncoder","action":"exclude"}'
func TestMarshalFuncEvenActionExclude(t *testing.T) {
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
