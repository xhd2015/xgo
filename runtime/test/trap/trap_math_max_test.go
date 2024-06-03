package trap

import (
	"math"
	"testing"
)

// these should compile
// see bug https://github.com/xhd2015/xgo/issues/53

const SET_EMPTY_STRING = "<empty>"
const SET_ZERO = math.MaxInt64 - 1
const SET_ZERO_UINT = math.MaxUint64 - 1
const SET_ZERO_INT32 = math.MaxInt32 - 1
const SET_ZERO_UINT32 = math.MaxUint32 - 1

func TestMathMaxShouldCompile(t *testing.T) {
	zero := uint64(SET_ZERO_UINT)
	if zero != math.MaxUint64-1 {
		t.Fatalf("zero: %v", zero)
	}
}
