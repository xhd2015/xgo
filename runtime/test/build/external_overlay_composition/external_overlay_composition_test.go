package external_overlay_composition

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestExternalOverlayIsComposedAfterInstrumentation is a regression for an
// external Go overlay whose replacement is subsequently instrumented by xgo.
// The final overlay emitted by xgo must keep the original source file as a
// key, mapping it to xgo's generated replacement.
func TestExternalOverlayIsComposedAfterInstrumentation(t *testing.T) {
	t.Parallel()

	output, err := runExternalOverlayFixture(t, "module")
	if err != nil {
		t.Fatalf("xgo must compose caller and generated overlays: %v\n%s", err, output)
	}
}

// runExternalOverlayFixture builds the named fixture through xgo with a caller
// overlay. Build-tagged compatibility tests use the same helper and differ
// only in the Go toolchain behavior they assert.
func runExternalOverlayFixture(t *testing.T, fixture string) ([]byte, error) {
	t.Helper()

	fixtureDir := t.TempDir()
	copyTree(t, filepath.Join("testdata", fixture), fixtureDir)

	original := filepath.Join(fixtureDir, "subject.go")
	replacement := filepath.Join(fixtureDir, "replacement", "subject.go")
	overlayFile := filepath.Join(fixtureDir, "caller-overlay.json")
	writeOverlay(t, overlayFile, original, replacement)

	args := []string{
		"run", "./cmd/xgo", "test",
		"--project-dir", fixtureDir,
		"-count=1",
		"-overlay", overlayFile,
		".",
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = xgoRepoRoot(t)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+filepath.Join(t.TempDir(), "go-cache"),
		"GOFLAGS=",
		"GOWORK=off",
	)
	return cmd.CombinedOutput()
}

func xgoRepoRoot(t *testing.T) string {
	t.Helper()
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return repoRoot
}

func writeOverlay(t *testing.T, file, original, replacement string) {
	t.Helper()
	data, err := json.Marshal(struct {
		Replace map[string]string
	}{Replace: map[string]string{original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func copyTree(t *testing.T, source, destination string) {
	t.Helper()
	if err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}); err != nil {
		t.Fatal(fmt.Errorf("copy fixture: %w", err))
	}
}
