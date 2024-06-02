package trap

import "testing"

var ints [3]int

// see bug https://github.com/xhd2015/xgo/issues/177
func TestNoTrapArrayPointer(t *testing.T) {
	x := &ints[0]
	y := &ints[0]

	if x != y {
		t.Fatalf("x != y: x=0x%x, y=0x%x", x, y)
	}
}
