package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/overlay"
)

func TestApplyBlankImportThroughOverlayInjectsIntoCallerReplacement(t *testing.T) {
	dir := t.TempDir()
	orig := filepath.Join(dir, "main.go")
	repl := filepath.Join(dir, "replacement.go")
	if err := os.WriteFile(orig, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(repl, []byte("package main\n\nfunc main() { println(\"overlay\") }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := overlay.MakeOverlay()
	src := overlay.AbsFile(orig)
	fs.OverrideFile(src, overlay.AbsFile(repl))

	// prep would have produced a copy of *orig* with the import; we only need
	// the source key so applyBlankImportThroughOverlay re-reads via overlay.
	prepTarget := overlay.AbsFile(filepath.Join(dir, "prep_main.go"))
	if err := applyBlankImportThroughOverlay(fs, map[overlay.AbsFile]overlay.AbsFile{
		src: prepTarget,
	}); err != nil {
		t.Fatal(err)
	}

	hit, content, err := fs.Read(src)
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("expected OverrideContent hit after blank-import inject")
	}
	if !strings.Contains(content, constants.RUNTIME_TRACE_PKG) {
		t.Fatalf("missing blank import in content:\n%s", content)
	}
	if !strings.Contains(content, `println("overlay")`) {
		t.Fatalf("caller replacement body lost:\n%s", content)
	}
	if strings.Contains(content, "func main() {}") && !strings.Contains(content, `println("overlay")`) {
		t.Fatalf("injected into original instead of replacement:\n%s", content)
	}
}

func TestApplyBlankImportThroughOverlaySkipsWhenAlreadyPresent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	body := "package main\n\nimport _ \"" + constants.RUNTIME_TRACE_PKG + "\"\n\nfunc main() {}\n"
	if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := overlay.MakeOverlay()
	src := overlay.AbsFile(file)
	if err := applyBlankImportThroughOverlay(fs, map[overlay.AbsFile]overlay.AbsFile{
		src: overlay.AbsFile(filepath.Join(dir, "unused.go")),
	}); err != nil {
		t.Fatal(err)
	}
	// No content override needed when import already present: Read should
	// still return the on-disk body (hitContent false).
	hit, content, err := fs.Read(src)
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("unexpected content override when import already present")
	}
	if !strings.Contains(content, constants.RUNTIME_TRACE_PKG) {
		t.Fatalf("disk content missing import:\n%s", content)
	}
}

func TestApplyBlankImportThroughOverlayNoCaller(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := overlay.MakeOverlay()
	src := overlay.AbsFile(file)
	if err := applyBlankImportThroughOverlay(fs, map[overlay.AbsFile]overlay.AbsFile{
		src: "",
	}); err != nil {
		t.Fatal(err)
	}
	hit, content, err := fs.Read(src)
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("expected content override without caller overlay")
	}
	if !strings.Contains(content, constants.RUNTIME_TRACE_PKG) {
		t.Fatalf("missing blank import:\n%s", content)
	}
}
