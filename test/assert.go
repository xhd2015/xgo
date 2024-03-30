package test

import (
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/strutil"
)

func containsLine(lines []string, line string) bool {
	for _, t := range lines {
		if t == line {
			return true
		}
	}
	return false
}

func expectContains(t *testing.T, output string, expectLines []string) {
	for _, expectLine := range expectLines {
		if !strings.Contains(output, expectLine) {
			t.Fatalf("expect output contains %q, actually not found", expectLine)
		}
	}
}

func expectSequence(t *testing.T, output string, sequence []string) {
	err := strutil.CheckSequence(output, sequence)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
