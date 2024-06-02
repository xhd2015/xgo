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

const userNsLength = (1 << 16)
const (
	minimumMappingUID = userNsLength
	mappingLen        = userNsLength * 2000
)

// see bug https://github.com/xhd2015/xgo/issues/176
func TestUntypedUnknownConstShouldCompile(t *testing.T) {
	var allocated uint32

	if allocated > minimumMappingUID+mappingLen-userNsLength {
		t.Fatalf("allocated is greater?")
	}
}
