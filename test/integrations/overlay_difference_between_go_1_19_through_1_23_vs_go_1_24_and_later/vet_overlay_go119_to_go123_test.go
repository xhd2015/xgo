//go:build go1.19 && !go1.24

package overlay_difference_between_go_1_19_through_1_23_vs_go_1_24_and_later

import (
	"strings"
	"testing"
)

func assertNativeGoVetOverlayOutcome(t *testing.T, output []byte, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Go 1.19 through 1.23 must reject the caller overlay during vet package loading:\n%s", output)
	}
	if !strings.Contains(string(output), "example.com/native-overlay-vet/not_present") {
		t.Fatalf("want Go 1.19 through 1.23 vet import failure, got:\n%s", output)
	}
}
