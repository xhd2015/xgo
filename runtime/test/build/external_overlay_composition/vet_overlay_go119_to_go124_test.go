//go:build go1.19 && !go1.25

package external_overlay_composition

import (
	"strings"
	"testing"
)

// Go 1.19 through 1.24 vet the original source instead of the caller's
// overlay replacement. Keep this known toolchain behavior explicit so a
// future xgo change does not mistake it for an overlay-composition failure.
func TestVetDoesNotApplyCallerOverlayBeforeGo125(t *testing.T) {
	t.Parallel()

	output, err := runExternalOverlayFixture(t, "vet_module")
	if err == nil {
		t.Fatalf("Go vet unexpectedly applied the caller overlay; output:\n%s", output)
	}
	if !strings.Contains(string(output), "could not import example.com/external-overlay-target/invalid") {
		t.Fatalf("want the known Go vet overlay failure, got:\n%s", output)
	}
}
