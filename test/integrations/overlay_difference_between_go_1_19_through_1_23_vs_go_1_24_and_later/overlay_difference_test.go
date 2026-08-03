//go:build go1.19

package overlay_difference_between_go_1_19_through_1_23_vs_go_1_24_and_later

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNativeGoVetOverlayImportResolution documents the Go 1.24 compatibility
// boundary for vet's package loading: Go 1.19 through 1.23 resolve imports
// from the original source, while Go 1.24 and later resolve them through the
// caller overlay. Version-specific assertions live in build-tagged files.
func TestNativeGoVetOverlayImportResolution(t *testing.T) {
	output, err := runNativeGoVetOverlayFixture(t)
	assertNativeGoVetOverlayOutcome(t, output, err)
}

func runNativeGoVetOverlayFixture(t *testing.T) ([]byte, error) {
	t.Helper()

	fixture := t.TempDir()
	copyTree(t, "testdata/module", fixture)
	overlayFile := filepath.Join(fixture, "caller-overlay.json")

	cmd := exec.Command("go", "test", "-count=1", "-overlay", overlayFile, ".")
	cmd.Dir = fixture
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(fixture, "go-cache"), "GOFLAGS=", "GOWORK=off")
	return cmd.CombinedOutput()
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
		t.Fatal(err)
	}
}
