package overlay_difference_between_go_1_19_through_1_23_vs_go_1_24_and_later

import (
	"encoding/json"
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
	writeFile(t, fixture, "go.mod", "module example.com/native-overlay-vet\n\ngo 1.19\n")
	writeFile(t, fixture, "subject.go", `package target

import _ "example.com/native-overlay-vet/not_present"

func Value() string {
	return "original"
}
`)
	writeFile(t, fixture, filepath.Join("replacement", "subject.go"), `package target

func Value() string { return "caller replacement" }
`)
	writeFile(t, fixture, "subject_test.go", `package target

import "testing"

func TestValue(t *testing.T) {
	if got, want := Value(), "caller replacement"; got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}
`)

	overlayFile := filepath.Join(fixture, "caller-overlay.json")
	overlayData, err := json.Marshal(struct {
		Replace map[string]string
	}{Replace: map[string]string{
		"subject.go": "replacement/subject.go",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overlayFile, overlayData, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "test", "-count=1", "-overlay", overlayFile, ".")
	cmd.Dir = fixture
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(fixture, "go-cache"), "GOFLAGS=", "GOWORK=off")
	return cmd.CombinedOutput()
}

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
