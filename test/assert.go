package test

import (
	"strings"
	"testing"
)

func containsLine(lines []string, line string) bool {
	for _, t := range lines {
		if t == line {
			return true
		}
	}
	return false
}

func expectSequence(t *testing.T, output string, sequence []string) {
	for i, s := range sequence {
		idx := strings.Index(output, s)
		if idx < 0 {
			t.Fatalf("expect output contains %q at sequence %d, actually not found", s, i)
		}
		output = output[idx+len(s):]
	}
}
