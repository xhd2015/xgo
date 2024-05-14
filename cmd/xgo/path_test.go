package main

import (
	"path/filepath"
	"runtime"
	"testing"
)

// go test -run TestFilePathDir -v ./cmd/xgo
func TestFilePathDir(t *testing.T) {

	type testCase struct {
		Name string
		Dir  string
	}
	var testCases = []testCase{
		{"", "."},
		{".", "."},
	}
	if runtime.GOOS != "windows" {
		testCases = append(testCases, []testCase{
			{"/", "/"},
			{"/tmp", "/"},
			{"//tmp", "/"},
		}...)
	} else {
		testCases = append(testCases, []testCase{
			{"/", "\\"},
			{"/tmp", "\\"},
			{"//tmp", "\\\\tmp"},
		}...)
	}

	for _, testCase := range testCases {
		res := filepath.Dir(testCase.Name)
		if res != testCase.Dir {
			t.Fatalf("expect dir(%q) = %q, actual: %q", testCase.Name, testCase.Dir, res)
		}
	}
}
