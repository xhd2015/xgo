package coverage

import "testing"

func TestParseCoverage(t *testing.T) {
	mode, lines := Parse(`mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 0`)
	if mode != "set" {
		t.Fatalf("expect mode to be %s, actual: %s", "set", mode)
	}
	if len(lines) != 1 {
		t.Fatalf("len(lines): %d", len(lines))
	}
	// t.Logf("lines[0]: %v", lines[0])
}
