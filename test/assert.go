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
	for _, s := range sequence {
		idx := strings.Index(output, s)
		if idx < 0 {
			t.Fatalf("expect output contains %q, actually not found", s)
		}
		output = output[idx+len(s):]
	}
}
