package testconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

func TestParseGoMinMaxObjectAndString(t *testing.T) {
	cfg, err := Parse([]byte(`{"go":{"min":"1.18","max":"1.20"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Go == nil || cfg.Go.Min != "1.18" || cfg.Go.Max != "1.20" {
		t.Fatalf("Go=%#v", cfg.Go)
	}

	cfg, err = Parse([]byte(`{"go":"1.19"}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Go == nil || cfg.Go.Min != "1.19" || cfg.Go.Max != "" {
		t.Fatalf("string go form: %#v", cfg.Go)
	}
}

func TestParseIgnoresUnknownKeys(t *testing.T) {
	cfg, err := Parse([]byte(`{
  "go":{"min":"1.18"},
  "pre_test":[{"command":["tool"]}],
  "flags":["--unified"],
  "env":{"A":true}
}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Go.Min != "1.18" || len(cfg.Flags) != 1 || cfg.Env["A"] != true {
		t.Fatalf("cfg=%#v", cfg)
	}
}

func TestParseMockRulesAndBypass(t *testing.T) {
	cfg, err := Parse([]byte(`{
  "bypass_go_flags": true,
  "args":["--config=x"],
  "mock_rules":[{"any":true,"action":"exclude"}]
}`))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.BypassGoFlags || len(cfg.Args) != 1 {
		t.Fatalf("args/bypass %#v", cfg)
	}
	if len(cfg.MockRules) != 1 || !strings.Contains(cfg.MockRules[0], `"any":true`) {
		t.Fatalf("MockRules=%v", cfg.MockRules)
	}
}

func TestLoadMissing(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil || cfg != nil {
		t.Fatalf("got (%v, %v)", cfg, err)
	}
}

func TestLoadEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultFileName)
	if err := os.WriteFile(path, []byte("  \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil || cfg == nil {
		t.Fatalf("got (%v, %v)", cfg, err)
	}
}

func TestCompareGoVersionIgnorePatch(t *testing.T) {
	a := &goinfo.GoVersion{Major: 1, Minor: 20, Patch: 5}
	b := &goinfo.GoVersion{Major: 1, Minor: 20, Patch: 0}
	if CompareGoVersion(a, b, true) != 0 {
		t.Fatal("ignore patch should equal")
	}
	if CompareGoVersion(a, b, false) <= 0 {
		t.Fatal("with patch a > b")
	}
	c := &goinfo.GoVersion{Major: 1, Minor: 21, Patch: 0}
	if CompareGoVersion(c, b, true) <= 0 {
		t.Fatal("1.21 > 1.20")
	}
}

func TestValidateGoConstraintNoOp(t *testing.T) {
	if err := ValidateGoConstraint(nil, "go"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGoConstraint(&GoConfig{}, "go"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateGoConstraintAgainstHost(t *testing.T) {
	// Wide range always accepts current host.
	if err := ValidateGoConstraint(&GoConfig{Min: "1.0", Max: "99.0"}, "go"); err != nil {
		t.Fatal(err)
	}
	// Impossible min.
	err := ValidateGoConstraint(&GoConfig{Min: "99.0"}, "go")
	if err == nil || !strings.Contains(err.Error(), "< 99.0") {
		t.Fatalf("want min error, got %v", err)
	}
	// Impossible max.
	err = ValidateGoConstraint(&GoConfig{Max: "1.0"}, "go")
	if err == nil || !strings.Contains(err.Error(), "> 1.0") {
		t.Fatalf("want max error, got %v", err)
	}
}

func TestValidateGoConstraintIgnoresBadToken(t *testing.T) {
	// Bad min is ignored; only max applies. max 99 always ok.
	if err := ValidateGoConstraint(&GoConfig{Min: "not-a-version", Max: "99.0"}, "go"); err != nil {
		t.Fatal(err)
	}
}
