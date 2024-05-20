package goinfo

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestFilePathDir(t *testing.T) {
	type Test struct {
		name string
		dir  string
		want string
	}
	tests := []Test{
		// NOTE: must convert dot path to abs path
		{"dot", ".", "."},
		{"dot_slash", "./", "."},
	}
	if runtime.GOOS != "windows" {
		tests = append(tests, Test{"root", "/", "/"})
	} else {
		tests = append(tests, Test{"root_windows", "/", "\\"})
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := filepath.Dir(tt.dir)
			if got != tt.want {
				t.Errorf("filepath.Dir(%q) = %q, want %q", tt.dir, got, tt.want)
			}
		})
	}
}
