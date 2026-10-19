package filecopy

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLinkReplaceDirAndUnshare(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("xgo does not hard-link GOROOTs on windows")
	}
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dst")
	file := filepath.Join(src, "a", "tool")
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("orig"), 0555); err != nil {
		t.Fatal(err)
	}
	if err := LinkReplaceDir(src, dst); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(dst, "a", "tool")
	si, _ := os.Stat(file)
	di, err := os.Stat(linked)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(si, di) {
		t.Fatalf("expected a hard link")
	}

	if err := Unshare(linked); err != nil {
		t.Fatal(err)
	}
	di, err = os.Stat(linked)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(si, di) {
		t.Fatalf("still shared after Unshare")
	}
	if runtime.GOOS != "windows" && di.Mode().Perm() != 0755 {
		t.Errorf("mode = %v, want the source mode plus owner write", di.Mode().Perm())
	}
	if err := os.WriteFile(linked, []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(file); string(got) != "orig" {
		t.Errorf("source modified: %q", got)
	}
	entries, _ := os.ReadDir(filepath.Join(dst, "a"))
	if len(entries) != 1 {
		t.Errorf("leftover temp files: %v", entries)
	}
	if err := Unshare(filepath.Join(dst, "missing")); err != nil {
		t.Errorf("Unshare of a missing file: %v", err)
	}
}

func TestUnshareFollowsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("xgo does not hard-link GOROOTs on windows")
	}
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dst")
	file := filepath.Join(src, "tool")
	if err := os.WriteFile(file, []byte("orig"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("tool", filepath.Join(src, "alias")); err != nil {
		t.Fatal(err)
	}
	if err := LinkReplaceDir(src, dst); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(dst, "alias")
	if err := Unshare(alias); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Lstat(alias); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink not kept: %v %v", fi, err)
	}
	if err := os.WriteFile(alias, []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(file); string(got) != "orig" {
		t.Errorf("source modified through the symlink: %q", got)
	}
	if err := os.Symlink("nowhere", filepath.Join(dst, "dangling")); err != nil {
		t.Fatal(err)
	}
	if err := Unshare(filepath.Join(dst, "dangling")); err != nil {
		t.Errorf("Unshare of a dangling symlink: %v", err)
	}
}
