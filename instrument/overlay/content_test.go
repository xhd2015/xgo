package overlay

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Runs generating different contents for a file must not overwrite each
// other's, and the same content must map to the same file.
func TestMakeGoOverlayKeepsContentsApart(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src", "a.go")
	genDir := filepath.Join(dir, "gen")
	target := func(content string) string {
		t.Helper()
		fs := MakeOverlay()
		fs.OverrideContent(AbsFile(src), content)
		goOverlay, err := fs.MakeGoOverlay(genDir, Options{NoLineDirective: true})
		if err != nil {
			t.Fatal(err)
		}
		return filepath.FromSlash(string(goOverlay.Replace[AbsFile(filepath.ToSlash(src))]))
	}
	one := target("package a // one\n")
	two := target("package a // two\n")
	if one == two {
		t.Fatalf("different contents share %s", one)
	}
	if filepath.Base(one) != "a.go" {
		t.Errorf("base name changed: %s", one)
	}
	for file, want := range map[string]string{one: "package a // one\n", two: "package a // two\n"} {
		data, err := os.ReadFile(file)
		if err != nil || string(data) != want {
			t.Errorf("%s: %q, %v; want %q", file, data, err, want)
		}
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(one, old, old); err != nil {
		t.Fatal(err)
	}
	if again := target("package a // one\n"); again != one {
		t.Fatalf("same content written to %s, then %s", one, again)
	}
	if info, err := os.Stat(one); err != nil || !info.ModTime().After(old) {
		t.Errorf("reused file not touched: %v %v", info.ModTime(), err)
	}
	if tmps, _ := filepath.Glob(filepath.Join(filepath.Dir(one), ".tmp-*")); len(tmps) > 0 {
		t.Errorf("temporary files left: %v", tmps)
	}
}

func TestCleanGoOverlay(t *testing.T) {
	dir := t.TempDir()
	stale, err := WriteContentFile(filepath.Join(dir, "PROJECT", "a"), "a.go", []byte("package a\n"))
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := WriteContentFile(filepath.Join(dir, "PROJECT", "a"), "b.go", []byte("package b\n"))
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, "PROJECT", "go.mod")
	if err := os.WriteFile(other, []byte("module m\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-8 * 24 * time.Hour)
	for _, f := range []string{stale, filepath.Dir(stale), other} {
		if err := os.Chtimes(f, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if err := CleanGoOverlay(dir, 7*24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale file kept: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(stale)); !os.IsNotExist(err) {
		t.Errorf("empty stale content dir kept: %v", err)
	}
	for _, f := range []string{fresh, other} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("%s removed: %v", f, err)
		}
	}
}
