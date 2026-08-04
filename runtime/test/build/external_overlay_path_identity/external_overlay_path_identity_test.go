package external_overlay_path_identity

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// MECE path_identity leaves: prove absFileKey-style unification under a real
// xgo run — absolute overlay JSON keys, and project-dir accessed via symlink
// (macOS /var vs /private/var class).

func TestAbsoluteOverlayKeysComposeAndMock(t *testing.T) {
	root := findXgoRepoRoot(t)
	modDir := writeSubjectModule(t)
	src := filepath.Join(modDir, "subject.go")
	repl := filepath.Join(modDir, "replacement", "subject.go")
	overlayPath := filepath.Join(modDir, "overlay-abs.json")
	writeOverlayJSON(t, overlayPath, map[string]string{
		// Absolute source/target as produced by some tooling / GenerateOverlay.
		src: repl,
	})
	runNestedXgoTest(t, root, modDir, overlayPath)
}

func TestSymlinkProjectDirComposeAndMock(t *testing.T) {
	root := findXgoRepoRoot(t)
	realDir := writeSubjectModule(t)
	linkParent := t.TempDir()
	linkDir := filepath.Join(linkParent, "proj-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink not available: %v", err)
	}
	// Prefer the non-eval spelling when it differs (macOS /var vs /private/var).
	projectDir := linkDir
	if eval, err := filepath.EvalSymlinks(linkDir); err == nil && eval != linkDir {
		// Invoke xgo with the link path; load may report eval'd AbsPaths.
		projectDir = linkDir
		t.Logf("project-dir link=%q eval=%q", linkDir, eval)
	}
	overlayPath := filepath.Join(projectDir, "caller-overlay.json")
	// Relative overlay resolved against project-dir (the symlink path).
	writeOverlayJSON(t, overlayPath, map[string]string{
		"subject.go": "replacement/subject.go",
	})
	runNestedXgoTest(t, root, projectDir, overlayPath)
}

func runNestedXgoTest(t *testing.T, xgoRoot, projectDir, overlayPath string) {
	t.Helper()
	args := []string{
		"run", "./cmd/xgo",
		"test",
		"--project-dir", projectDir,
		"-overlay", overlayPath,
		"--mock-rule", `{"main_module":true,"kind":"func","action":"include"}`,
		"--mock-rule", `{"any":true,"action":"exclude"}`,
		"-count=1",
		"-timeout=60s",
		".",
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = xgoRoot
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		t.Fatalf("nested xgo test failed: %v\n%s", err, buf.String())
	}
	// Success is exit 0 from nested `xgo test` (replacement content + mock).
	if t.Failed() {
		return
	}
	t.Logf("nested xgo output:\n%s", buf.String())
}

func writeSubjectModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Force TMPDIR-style paths on macOS when possible (often /var/folders/…).
	mustWrite(t, filepath.Join(dir, "go.mod"), ""+
		"module example.com/overlay-path-identity\n\n"+
		"go 1.18\n\n"+
		"require github.com/xhd2015/xgo/runtime v0.0.0\n\n"+
		"replace github.com/xhd2015/xgo/runtime => "+runtimeReplacePath(t)+"\n",
	)
	mustWrite(t, filepath.Join(dir, "subject.go"), ""+
		"package pathid\n\n"+
		"func Value() string { return \"original\" }\n",
	)
	mustWrite(t, filepath.Join(dir, "replacement", "subject.go"), ""+
		"package pathid\n\n"+
		"func Value() string { return \"caller replacement\" }\n",
	)
	mustWrite(t, filepath.Join(dir, "pathid_test.go"), ""+
		"package pathid\n\n"+
		"import (\n"+
		"\t\"testing\"\n\n"+
		"\t\"github.com/xhd2015/xgo/runtime/mock\"\n"+
		")\n\n"+
		"func TestValueAndMock(t *testing.T) {\n"+
		"\tif got, want := Value(), \"caller replacement\"; got != want {\n"+
		"\t\tt.Fatalf(\"Value() = %q, want %q\", got, want)\n"+
		"\t}\n"+
		"\tmock.Patch(Value, func() string { return \"mocked\" })\n"+
		"\tif got, want := Value(), \"mocked\"; got != want {\n"+
		"\t\tt.Fatalf(\"after mock Value() = %q, want %q\", got, want)\n"+
		"\t}\n"+
		"}\n",
	)
	return dir
}

func writeOverlayJSON(t *testing.T, path string, replace map[string]string) {
	t.Helper()
	// Normalize to slash form like Go overlay JSON often uses.
	norm := make(map[string]string, len(replace))
	for k, v := range replace {
		norm[filepath.ToSlash(k)] = filepath.ToSlash(v)
	}
	body, err := json.Marshal(struct {
		Replace map[string]string `json:"Replace"`
	}{Replace: norm})
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, path, string(body))
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runtimeReplacePath(t *testing.T) string {
	t.Helper()
	root := findXgoRepoRoot(t)
	// go.mod replace must be absolute or relative to the temp module; use absolute.
	p := filepath.Join(root, "runtime")
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(abs)
}

func isXgoMainModule(goMod []byte) bool {
	// Exact main module path only — not runtime/test/... submodules.
	for _, line := range bytes.Split(goMod, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if bytes.Equal(line, []byte("module github.com/xhd2015/xgo")) {
			return true
		}
	}
	return false
}

func findXgoRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && isXgoMainModule(data) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Fallback: runtime/test/build/external_overlay_path_identity -> repo root.
			cand := filepath.Join(wd, "..", "..", "..", "..")
			if abs, err := filepath.Abs(cand); err == nil {
				if data, err := os.ReadFile(filepath.Join(abs, "go.mod")); err == nil && isXgoMainModule(data) {
					return abs
				}
			}
			t.Skipf("xgo repo root not found from %s (GOOS=%s)", wd, runtime.GOOS)
		}
		dir = parent
	}
}
