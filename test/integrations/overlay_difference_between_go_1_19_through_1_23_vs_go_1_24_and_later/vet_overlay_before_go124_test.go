//go:build !go1.24

package overlay_difference_between_go_1_19_through_1_23_vs_go_1_24_and_later

import "testing"

func assertNativeGoVetOverlayOutcome(t *testing.T, output []byte, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Go versions before 1.24 must reject the caller overlay during vet package loading:\n%s", output)
	}
}
