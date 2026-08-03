package overlay_difference_between_go_1_19_through_1_23_vs_go_1_24_and_later

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Native Go does not apply this caller overlay while vetting. This is kept as
// a separate integration check from xgo's composition behavior: all supported
// Go versions must reject the vet error in the original source.
func TestNativeGoVetUsesOriginalSourceDespiteCallerOverlay(t *testing.T) {
	fixture := t.TempDir()
	writeFile(t, fixture, "go.mod", "module example.com/native-overlay-vet\n\ngo 1.19\n")
	writeFile(t, fixture, "subject.go", `package target

import "fmt"

func Value() string {
	fmt.Printf("%d", "not a number")
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
		filepath.Join(fixture, "subject.go"): filepath.Join(fixture, "replacement", "subject.go"),
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
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("native go test unexpectedly accepted caller overlay:\n%s", output)
	}
	if !strings.Contains(string(output), "fmt.Printf format %d") {
		t.Fatalf("want vet error from original source, got:\n%s", output)
	}
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
