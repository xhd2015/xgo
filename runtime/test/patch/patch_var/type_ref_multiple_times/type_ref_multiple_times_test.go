package type_ref_multiple_times

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/patch/patch_var/type_ref_multiple_times/types"
)

var testMap = map[types.Key]types.Value{
	"A": "1",
}

func TestTypeRefMultipleTimes(t *testing.T) {
	res := testMap["A"]
	if res != "1" {
		t.Errorf("expect testMap[\"A\"] = %v, but got %v", "1", res)
	}
	mock.Patch(&testMap, func() map[types.Key]types.Value {
		return map[types.Key]types.Value{
			"A": "mock",
		}
	})
	res = testMap["A"]
	if res != "mock" {
		t.Errorf("expect testMap[\"A\"] = %v, but got %v", "mock", res)
	}
}
