//go:build go1.24

package overlay_difference_between_go_1_19_through_1_23_vs_go_1_24_and_later

import "testing"

func assertNativeGoVetOverlayOutcome(t *testing.T, output []byte, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Go 1.24 and later must resolve imports through the caller overlay: %v\n%s", err, output)
	}
}
