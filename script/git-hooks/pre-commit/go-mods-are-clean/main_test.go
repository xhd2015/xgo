package main

import (
	"strings"
	"testing"
)

func TestCheckGoModContent_Clean(t *testing.T) {
	cases := []string{
		"module github.com/xhd2015/xgo\n\ngo 1.18\n",
		"module github.com/xhd2015/xgo/runtime\n\ngo 1.14\n",
		"module foo\n\n",
		"module foo\ngo 1.21\n",
		"module foo\ngo 1.21\ntoolchain go1.22.0\n",
	}

	for i, content := range cases {
		err := checkGoModContent("go.mod", content)
		if err != nil {
			t.Errorf("case %d: expected clean, got error: %v\ncontent:\n%s", i, err, content)
		}
	}
}

func TestCheckGoModContent_Dirty_Require_SingleLine(t *testing.T) {
	content := "module github.com/xhd2015/xgo\n\ngo 1.18\n\nrequire github.com/foo/bar v1.0.0\n"
	err := checkGoModContent("go.mod", content)
	if err == nil {
		t.Fatal("expected error for require directive")
	}
	if !strings.Contains(err.Error(), "require") {
		t.Fatalf("error should mention require, got: %v", err)
	}
}

func TestCheckGoModContent_Dirty_Require_Block(t *testing.T) {
	content := `module github.com/xhd2015/xgo

go 1.18

require (
	github.com/foo/bar v1.0.0
)
`
	err := checkGoModContent("go.mod", content)
	if err == nil {
		t.Fatal("expected error for require block directive")
	}
	if !strings.Contains(err.Error(), "require") {
		t.Fatalf("error should mention require, got: %v", err)
	}
}

func TestCheckGoModContent_Dirty_Replace(t *testing.T) {
	content := `module github.com/xhd2015/xgo

go 1.18

replace github.com/foo/bar => ../bar
`
	err := checkGoModContent("go.mod", content)
	if err == nil {
		t.Fatal("expected error for replace directive")
	}
	if !strings.Contains(err.Error(), "replace") {
		t.Fatalf("error should mention replace, got: %v", err)
	}
}

func TestCheckGoModContent_Dirty_Exclude(t *testing.T) {
	content := `module github.com/xhd2015/xgo

go 1.18

exclude github.com/foo/bar v1.0.0
`
	err := checkGoModContent("go.mod", content)
	if err == nil {
		t.Fatal("expected error for exclude directive")
	}
	if !strings.Contains(err.Error(), "exclude") {
		t.Fatalf("error should mention exclude, got: %v", err)
	}
}

func TestCheckGoModContent_Dirty_Retract(t *testing.T) {
	content := `module github.com/xhd2015/xgo

go 1.18

retract (
	v1.0.0
)
`
	err := checkGoModContent("go.mod", content)
	if err == nil {
		t.Fatal("expected error for retract directive")
	}
	if !strings.Contains(err.Error(), "retract") {
		t.Fatalf("error should mention retract, got: %v", err)
	}
}

func TestCheckGoModContent_Empty(t *testing.T) {
	err := checkGoModContent("go.mod", "")
	if err != nil {
		t.Fatalf("empty file should be clean, got: %v", err)
	}
}

func TestCheckGoModContent_OnlyWhitespace(t *testing.T) {
	err := checkGoModContent("go.mod", "\n\n  \n\t\n")
	if err != nil {
		t.Fatalf("whitespace-only file should be clean, got: %v", err)
	}
}
