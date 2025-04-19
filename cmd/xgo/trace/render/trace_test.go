package render

import (
	"testing"
)

// table:
//
//	cost result c
//	10 10ns
//	10342 10us 30
func TestFormatCost(t *testing.T) {
	n := formatCost(0, 64)
	if n != "64ns" {
		t.Fatalf("expect %s, actual: %s", "64ns", n)
	}

	n = formatCost(0, 64103)
	if n != "64μs" {
		t.Fatalf("expect %s, actual: %s", "64us", n)
	}
}
