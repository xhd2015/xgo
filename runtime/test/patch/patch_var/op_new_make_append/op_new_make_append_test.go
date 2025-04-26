package op_new_make

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

var A = make(map[string]int)
var B = new(map[string]int)
var C = append([]int{}, 1, 2, 3)

func TestOpMake(t *testing.T) {
	before := A["hello"]
	if before != 0 {
		t.Errorf("expect before: %d, actual: %d", 0, before)
	}
	mock.Patch(&A, func() map[string]int {
		return map[string]int{"hello": 123}
	})
	after := A["hello"]
	if after != 123 {
		t.Errorf("expect after: %d, actual: %d", 123, after)
	}
}

func TestOpNew(t *testing.T) {
	before := (*B)["hello"]
	if before != 0 {
		t.Errorf("expect before: %d, actual: %d", 0, before)
	}
	mock.Patch(&B, func() *map[string]int {
		return &map[string]int{"hello": 123}
	})
	after := (*B)["hello"]
	if after != 123 {
		t.Errorf("expect after: %d, actual: %d", 123, after)
	}
}

func TestOpAppend(t *testing.T) {
	before := C
	if before[0] != 1 || before[1] != 2 || before[2] != 3 {
		t.Errorf("expect before: %v, actual: %v", []int{1, 2, 3}, before)
	}
	mock.Patch(&C, func() []int {
		return []int{4, 5, 6}
	})
	after := C
	if after[0] != 4 || after[1] != 5 || after[2] != 6 {
		t.Errorf("expect after: %v, actual: %v", []int{4, 5, 6}, after)
	}
}

func TestDebugOpAppend(t *testing.T) {
	b := C
	_ = b
}
