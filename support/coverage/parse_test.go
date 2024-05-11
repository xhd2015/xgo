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

func TestCompact(t *testing.T) {
	mode, lines := Parse(`mode: set
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 1
github.com/xhd2015/xgo/runtime/core/func1.go:44.41,45.22 1 112
github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1 2`)
	if mode != "set" {
		t.Fatalf("expect mode to be %s, actual: %s", "set", mode)
	}
	if len(lines) != 3 {
		t.Fatalf("len(lines): %d", len(lines))
	}
	cmpLines := Compact(lines)
	if len(cmpLines) != 2 {
		t.Fatalf("len(cmpLines) : %d", len(cmpLines))
	}
	if cmpLines[0].Prefix != "github.com/xhd2015/xgo/runtime/core/func.go:44.41,45.22 1" {
		t.Fatalf("cmpLines[0].Prefix: %s", cmpLines[0].Prefix)
	}
	if cmpLines[0].Count != 3 {
		t.Fatalf("cmpLines[0].Count: %d", cmpLines[0].Count)
	}
	if cmpLines[1].Prefix != "github.com/xhd2015/xgo/runtime/core/func1.go:44.41,45.22 1" {
		t.Fatalf("cmpLines[1].Prefix: %s", cmpLines[1].Prefix)
	}
	if cmpLines[1].Count != 112 {
		t.Fatalf("cmpLines[1].Count: %d", cmpLines[1].Count)
	}
}
