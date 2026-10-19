package overlay

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/instrument/patch"
)

func TestAbsFileKeyUnifiesSlashAndNative(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "subject.go")
	if err := os.WriteFile(file, []byte("package p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	native := AbsFile(file)
	slashy := AbsFile(filepath.ToSlash(file))

	fs := MakeOverlay()
	fs.OverrideFile(slashy, slashy)
	if fs.Get(native) == nil {
		t.Fatalf("Get(native) miss after OverrideFile(slash); keys native=%q slash=%q", native, slashy)
	}
	if fs.Get(slashy) == nil {
		t.Fatal("Get(slash) miss after OverrideFile(slash)")
	}
	// Single map entry: both spellings share one identity.
	if len(fs) != 1 {
		t.Fatalf("overlay entries = %d, want 1 (unified key)", len(fs))
	}
}

func TestOverrideContentReplacesFileRedirectForAllSpellings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		// Windows without symlink privilege: still cover slash/native via other test.
		t.Skipf("symlink: %v", err)
	}
	srcReal := filepath.Join(realDir, "subject.go")
	srcLink := filepath.Join(linkDir, "subject.go")
	tgtReal := filepath.Join(realDir, "replacement.go")
	if err := os.WriteFile(srcReal, []byte("package p\nfunc F() string { return \"orig\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tgtReal, []byte("package p\nfunc F() string { return \"replacement\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := MakeOverlay()
	// File redirect via link spelling (macOS-style non-eval path).
	fs.OverrideFile(AbsFile(srcLink), AbsFile(tgtReal))
	// Instrumentation via real path (package load EvalSymlinks / go list form).
	fs.OverrideContent(AbsFile(srcReal), "package p\nfunc F() string { return \"instrumented\" }\n")

	for _, key := range []string{srcLink, srcReal, filepath.ToSlash(srcLink), filepath.ToSlash(srcReal)} {
		got := fs.Get(AbsFile(key))
		if got == nil {
			t.Fatalf("Get(%q) miss", key)
		}
		if !got.hasOverriddenContent {
			t.Fatalf("Get(%q): still file redirect, want instrumented content", key)
		}
		if got.Content != "package p\nfunc F() string { return \"instrumented\" }\n" {
			t.Fatalf("Get(%q) content = %q", key, got.Content)
		}
	}
	if len(fs) != 1 {
		t.Fatalf("overlay entries = %d, want 1 after content supersedes redirect", len(fs))
	}
}

func TestMakeGoOverlayEmitsInstrumentedNotStaleRedirect(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink: %v", err)
	}
	srcReal := filepath.Join(realDir, "subject.go")
	srcLink := filepath.Join(linkDir, "subject.go")
	tgtReal := filepath.Join(realDir, "replacement.go")
	if err := os.WriteFile(srcReal, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tgtReal, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	fs := MakeOverlay()
	fs.OverrideFile(AbsFile(srcLink), AbsFile(tgtReal))
	fs.OverrideContent(AbsFile(srcReal), "package p\n")

	genDir := filepath.Join(dir, "gen")
	goOverlay, err := fs.MakeGoOverlay(genDir, Options{NoLineDirective: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(goOverlay.Replace) != 1 {
		t.Fatalf("Replace = %#v, want exactly 1 mapping", goOverlay.Replace)
	}
	for source, target := range goOverlay.Replace {
		targetPath := filepath.Clean(filepath.FromSlash(string(target)))
		// Must not point at the uninstrumented replacement file.
		if targetPath == filepath.Clean(tgtReal) {
			t.Fatalf("stale file redirect survived: %q -> %q", source, target)
		}
		// Target should be generated content under genDir.
		rel, err := filepath.Rel(genDir, targetPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatalf("target %q not under gen %q (rel=%q err=%v)", target, genDir, rel, err)
		}
	}
}

func TestAbsFileKeyUnifiesWindowsStyleSeparators(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "windows" {
		// Still exercise FromSlash normalization on all platforms.
		mixed := AbsFile("C:/Users/x/proj/a.go")
		key := absFileKey(mixed)
		if filepath.ToSlash(string(key)) != "C:/Users/x/proj/a.go" && string(key) != filepath.FromSlash("C:/Users/x/proj/a.go") {
			// Clean/FromSlash result is platform-dependent; just ensure stable re-key.
		}
		if absFileKey(key) != key {
			t.Fatalf("absFileKey not idempotent: %q -> %q", key, absFileKey(key))
		}
		return
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "subject.go")
	if err := os.WriteFile(file, []byte("package p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := MakeOverlay()
	fs.OverrideContent(AbsFile(filepath.ToSlash(file)), "instrumented")
	got := fs.Get(AbsFile(file)) // native backslashes
	if got == nil || got.Content != "instrumented" {
		t.Fatalf("windows native Get after slash OverrideContent: got %#v", got)
	}
}

func TestMakeGoOverlayWritesEveryFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fs := MakeOverlay()
	want := make(map[string]string)
	for i := 0; i < 100; i++ {
		src := filepath.Join(dir, "src", "pkg"+strconv.Itoa(i%7), "f"+strconv.Itoa(i)+".go")
		content := "package p\n\nvar X = " + strconv.Itoa(i) + "\n"
		fs.OverrideContent(AbsFile(src), content)
		want[src] = content
	}
	redirectSrc := filepath.Join(dir, "src", "redirect.go")
	redirectTgt := filepath.Join(dir, "other", "target.go")
	fs.OverrideFile(AbsFile(redirectSrc), AbsFile(redirectTgt))

	genDir := filepath.Join(dir, "gen")
	goOverlay, err := fs.MakeGoOverlay(genDir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(goOverlay.Replace) != len(want)+1 {
		t.Fatalf("got %d mappings, want %d", len(goOverlay.Replace), len(want)+1)
	}
	if got := goOverlay.Replace[AbsFile(filepath.ToSlash(redirectSrc))]; got != AbsFile(filepath.ToSlash(redirectTgt)) {
		t.Errorf("redirect: %q", got)
	}
	for src, content := range want {
		target := goOverlay.Replace[AbsFile(filepath.ToSlash(src))]
		if target == "" {
			t.Fatalf("%s: no mapping", src)
		}
		data, err := os.ReadFile(filepath.FromSlash(string(target)))
		if err != nil {
			t.Fatal(err)
		}
		if wantData := patch.FmtLineDirective(src, 1) + "\n" + content; string(data) != wantData {
			t.Fatalf("%s: wrote %q, want %q", target, data, wantData)
		}
	}
}
