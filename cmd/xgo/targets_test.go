package main

import (
	"reflect"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

func TestTargetPatterns(t *testing.T) {
	tests := []struct {
		cmd  string
		args []string
		want []string
	}{
		{"test", nil, []string{"."}},
		{"test", []string{"-short", "./a", "./b/..."}, []string{"./a", "./b/..."}},
		{"test", []string{"./a", "-run", "TestX", "-v"}, []string{"./a"}},
		{"test", []string{"-count=1", "-p", "2", "./a"}, []string{"./a"}},
		{"test", []string{"--race", "-test.short", "./a"}, []string{"./a"}},
		{"build", []string{"-o", "out", "./cmd/x"}, []string{"./cmd/x"}},
		{"run", []string{"./cmd/x", "arg", "-flag"}, []string{"./cmd/x"}},
		{"run", []string{"main.go"}, nil},
		// an unknown flag may or may not take the next argument
		{"test", []string{"-somenewflag", "./a"}, nil},
	}
	for _, tt := range tests {
		if got := targetPatterns(tt.cmd, tt.args); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("targetPatterns(%q, %q) = %q, want %q", tt.cmd, tt.args, got, tt.want)
		}
	}
}

func TestMatchedPackages(t *testing.T) {
	list := []*goinfo.Package{
		{ImportPath: "m/a", Match: []string{"./a"}},
		{ImportPath: "m/b", Match: []string{"./b/..."}},
		{ImportPath: "m/b/c", Match: []string{"./b/..."}},
		{ImportPath: "m/d", DepOnly: true},
	}
	want := map[string]bool{"m/a": true, "m/b": true, "m/b/c": true}
	for _, patterns := range [][]string{
		{"./a", "./b/..."},
		// go list reports patterns cleaned
		{"./a/", ".//b/./..."},
	} {
		got := matchedPackages(list, patterns)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("matchedPackages(%q) = %v, want %v", patterns, got, want)
		}
	}
	// a pattern no package reports: everything is instrumented
	if got := matchedPackages(list, []string{"./a", "a/../x"}); got != nil {
		t.Errorf("unmatched pattern: got %v, want nil", got)
	}
}
