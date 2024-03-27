package main

import (
	"path/filepath"
	"testing"
)

// go test -run TestFilePathDir -v ./cmd/xgo
func TestFilePathDir(t *testing.T) {
	var testCases = []struct {
		Name string
		Dir  string
	}{
		{"", "."},
		{".", "."},
		{"/", "/"},
		{"/tmp", "/"},
		{"//tmp", "/"},
	}

	for _, testCase := range testCases {
		res := filepath.Dir(testCase.Name)
		if res != testCase.Dir {
			t.Fatalf("expect dir(%q) = %q, actual: %q", testCase.Name, testCase.Dir, res)
		}
	}
}
