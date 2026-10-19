package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCanonicalXgoHomeResolvesSymlinks(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "shared-xgo")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(root, "slot-home")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(home, ".xgo")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	got, err := getOrMakeAbsXgoHome(link)
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	defaultHome, err := getOrMakeAbsXgoHome("")
	if err != nil {
		t.Fatal(err)
	}
	if defaultHome != want {
		t.Fatalf("default xgo home = %q, want %q", defaultHome, want)
	}
	if got != want {
		t.Fatalf("canonicalXgoHome(%q) = %q, want %q", link, got, want)
	}
	missing := filepath.Join(root, "does-not-exist", ".xgo")
	if canonicalXgoHome(missing) != missing {
		t.Fatal("a missing path must be returned unchanged")
	}
}
