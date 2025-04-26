package op_new_make

import (
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/mock"
)

var A = 10 * 20
var B = 0x1 & 0x10

var C = 10*20 > 20

// see https://github.com/xhd2015/xgo/issues/331#issuecomment-2831811308
var D = 10 * time.Second

var F = 10 % time.Second //  time.Duration

var E = 10 ^ time.Second //  time.Duration

var G = 10 >> time.Nanosecond // int

func TestMul(t *testing.T) {
	before := A
	if before != 200 {
		t.Errorf("expect before: %d, actual: %d", 200, before)
	}
	mock.Patch(&A, func() int {
		return 123
	})
	after := A
	if after != 123 {
		t.Errorf("expect after: %d, actual: %d", 123, after)
	}
}

func TestAnd(t *testing.T) {
	before := B
	if before != 0 {
		t.Errorf("expect before: %d, actual: %d", 0, before)
	}
	mock.Patch(&B, func() int {
		return 123
	})
	after := B
	if after != 123 {
		t.Errorf("expect after: %d, actual: %d", 123, after)
	}
}
func TestMulLogic(t *testing.T) {
	before := C
	if before != true {
		t.Errorf("expect before: %t, actual: %t", true, before)
	}
	mock.Patch(&C, func() bool {
		return false
	})
	after := C
	if after != false {
		t.Errorf("expect after: %t, actual: %t", false, after)
	}
}

func TestMulTime(t *testing.T) {
	before := D
	if before != 10*time.Second {
		t.Errorf("expect before: %d, actual: %d", 10*time.Second, before)
	}
	mock.Patch(&D, func() time.Duration {
		return 123 * time.Second
	})
	after := D
	if after != 123*time.Second {
		t.Errorf("expect after: %d, actual: %d", 123*time.Second, after)
	}
}
