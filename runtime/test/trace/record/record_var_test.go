package record

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/trace"
)

var V int

func TestRecordVar(t *testing.T) {
	var records []string
	trace.RecordCall(&V, func(res *int) {
		records = append(records, fmt.Sprintf("V is accessed: %d", *res))
	})
	if V != 0 {
		t.Errorf("want V to be 0, but got %d", V)
	}
	V = 10
	if V != 10 {
		t.Errorf("want V to be 10, but got %d", V)
	}
	if len(records) != 2 {
		// the 2 records comes from read from `V!=0` and `V!=10`
		t.Errorf("records length is not 2: %d", len(records))
	}
	if len(records) > 0 && records[0] != "V is accessed: 0" {
		t.Errorf("records[0] is not 'V is accessed: 0': %s", records[0])
	}
	if len(records) > 1 && records[1] != "V is accessed: 10" {
		t.Errorf("records[1] is not 'V is accessed: 10': %s", records[1])
	}
}
