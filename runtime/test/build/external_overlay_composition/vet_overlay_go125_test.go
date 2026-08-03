//go:build go1.25

package external_overlay_composition

import "testing"

// Go 1.25 and newer apply the caller overlay while vetting this fixture.
func TestVetAppliesCallerOverlayAtGo125(t *testing.T) {
	t.Parallel()

	output, err := runExternalOverlayFixture(t, "vet_module")
	if err != nil {
		t.Fatalf("Go vet must apply the caller overlay: %v\n%s", err, output)
	}
}
