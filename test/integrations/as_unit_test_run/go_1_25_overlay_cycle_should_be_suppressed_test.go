//go:build go1.25

package as_unit_test_run

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGo1_25OverlayCycleShouldBeSuppressed(t *testing.T) {
	root, err := resolveRootDir()
	if err != nil {
		t.Fatal(err)
	}
	modDir := filepath.Join(root, "test", "integrations", "stricker_go_sum_policy_starting_go_1_25", "test_overlay_import_cycle")

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = modDir
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected import cycle error but test passed\noutput: %s", output)
	}

	errOutput := strings.ToLower(string(output))
	if strings.Contains(errOutput, "import cycle") ||
		strings.Contains(errOutput, "updates to go.mod needed") {
		t.Logf("import cycle correctly detected: %s", output)
		return
	}

	t.Fatalf("unexpected error: %v\noutput: %s", err, output)
}
