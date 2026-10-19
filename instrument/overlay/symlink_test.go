package overlay

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEvalSymlinksMatchesFilepath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("evalSymlinks is filepath.EvalSymlinks on windows")
	}
	root := t.TempDir()
	mustMkdir := func(p string) {
		if err := os.MkdirAll(filepath.Join(root, p), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite := func(p string) {
		if err := os.WriteFile(filepath.Join(root, p), []byte("package a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustLink := func(target, p string) {
		if err := os.Symlink(target, filepath.Join(root, p)); err != nil {
			t.Fatal(err)
		}
	}
	mustMkdir("real/sub")
	mustWrite("real/sub/a.go")
	mustWrite("real/b.go")
	mustLink("real", "dirlink")
	mustLink(filepath.Join(root, "real", "sub", "a.go"), "real/filelink.go")
	mustLink("sub/a.go", "real/rellink.go")
	mustLink("missing.go", "real/dangling.go")
	// directory links resolved through links, and relative to their parent
	mustLink("../../real", "real/sub/up")
	mustLink("dirlink/sub", "sublink")
	mustMkdir("deep/a/b")
	mustWrite("deep/a/b/c.go")
	mustLink("deep/a", "alink")

	paths := []string{
		"real/sub/a.go",
		"dirlink/sub/a.go",
		"dirlink/b.go",
		"dirlink/filelink.go",
		"dirlink/rellink.go",
		"dirlink/dangling.go",
		"dirlink/nofile.go",
		"nodir/x.go",
		"real/b.go/x.go",
		"dirlink/sub/../b.go",
		"dirlink",
		"real/sub/up/b.go",
		"dirlink/sub/up/sub/up/b.go",
		"sublink/a.go",
		"sublink/up/sub/a.go",
		"alink/b/c.go",
		"alink/b/missing/c.go",
		"deep/a/b/c.go",
	}
	// twice: resolving from the cache must not change the result
	for i := 0; i < 2; i++ {
		for _, p := range paths {
			path := filepath.Join(root, p)
			want, wantErr := filepath.EvalSymlinks(path)
			got, err := evalSymlinks(path)
			if (err != nil) != (wantErr != nil) || got != want {
				t.Errorf("evalSymlinks(%s) = %q, %v; filepath.EvalSymlinks: %q, %v", p, got, err, want, wantErr)
			}
		}
	}
	// a directory created after a failed lookup resolves
	if _, err := evalSymlinks(filepath.Join(root, "later", "c.go")); err == nil {
		t.Fatalf("expected an error before later/ exists")
	}
	mustMkdir("later")
	mustWrite("later/c.go")
	if got, err := evalSymlinks(filepath.Join(root, "later", "c.go")); err != nil || got != filepath.Join(root, "later", "c.go") {
		t.Errorf("created directory: %q, %v", got, err)
	}
}
