package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/osinfo"
)

// A repeated linked sync must only relink src/, keeping tools that setup
// rebuilt (they are not rebuilt again with --skip-rebuild-compiler-and-go).
func TestSyncGorootLinkedResyncKeepsTools(t *testing.T) {
	goroot := t.TempDir()
	inst := filepath.Join(t.TempDir(), "goroot")
	record := filepath.Join(t.TempDir(), "full-sync-record.txt")
	goBin := filepath.Join("bin", "go"+osinfo.EXE_SUFFIX)
	for name, content := range map[string]string{
		goBin: "original go",
		filepath.Join("src", "runtime", "proc.go"): "package runtime",
	} {
		p := filepath.Join(goroot, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := syncGoroot(goroot, inst, record, true); err != nil {
		t.Fatal(err)
	}
	instGo := filepath.Join(inst, goBin)
	if err := filecopy.Unshare(instGo); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(instGo, []byte("rebuilt go"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := syncGoroot(goroot, inst, record, true); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(instGo); string(got) != "rebuilt go" {
		t.Errorf("rebuilt go replaced on resync: %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(goroot, goBin)); string(got) != "original go" {
		t.Errorf("original go modified: %q", got)
	}
	if _, err := os.Stat(filepath.Join(inst, "src", "runtime", "proc.go")); err != nil {
		t.Errorf("src not synced: %v", err)
	}
}
