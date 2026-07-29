package main

import (
	"os"
	"testing"
)

func TestParseModuleList(t *testing.T) {
	got := parseModuleList(" a , b,a, ")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseOptions_IncludeAsMainModule_Flag(t *testing.T) {
	opts, err := parseOptions("test", []string{
		"--mock-rule-include-as-main-module=example.com/app,example.com/lib",
		"./...",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opts.mockRuleIncludeAsMainModule != "example.com/app,example.com/lib" {
		t.Fatalf("flag effective: %q", opts.mockRuleIncludeAsMainModule)
	}
}

func TestParseOptions_IncludeAsMainModule_EnvFallback(t *testing.T) {
	os.Setenv("XGO_MOCK_RULE_INCLUDE_AS_MAIN_MODULE", "example.com/fromenv")
	defer os.Unsetenv("XGO_MOCK_RULE_INCLUDE_AS_MAIN_MODULE")
	opts, err := parseOptions("test", []string{"./..."})
	if err != nil {
		t.Fatal(err)
	}
	if opts.mockRuleIncludeAsMainModule != "example.com/fromenv" {
		t.Fatalf("env effective: %q", opts.mockRuleIncludeAsMainModule)
	}
}

func TestParseOptions_IncludeAsMainModule_FlagWinsEnv(t *testing.T) {
	os.Setenv("XGO_MOCK_RULE_INCLUDE_AS_MAIN_MODULE", "example.com/fromenv")
	defer os.Unsetenv("XGO_MOCK_RULE_INCLUDE_AS_MAIN_MODULE")
	opts, err := parseOptions("test", []string{
		"--mock-rule-include-as-main-module=example.com/fromflag",
		"./...",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opts.mockRuleIncludeAsMainModule != "example.com/fromflag" {
		t.Fatalf("flag should win: %q", opts.mockRuleIncludeAsMainModule)
	}
}

func TestPkgWithinAnyModule(t *testing.T) {
	if !pkgWithinAnyModule("example.com/app/foo", "example.com/suite", []string{"example.com/app"}) {
		t.Fatal("expected app/foo under app")
	}
	if pkgWithinAnyModule("example.com/other/x", "example.com/suite", []string{"example.com/app"}) {
		t.Fatal("other should not match")
	}
}
