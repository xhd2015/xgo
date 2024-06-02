// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"testing"
)

var ints [3]int

func TestArrayPointer(t *testing.T) {

	x := &ints[0]
	y := &ints[0]

	t.Logf("x=0x%x", x)
	t.Logf("y=0x%x", y)
	if x != y {
		t.Fatalf("x != y")
	}
}
