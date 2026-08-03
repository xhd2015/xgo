//go:build !go1.19

package overlay_difference_between_go_1_19_through_1_23_vs_go_1_24_and_later

import "testing"

func TestNativeGoVetOverlayImportResolutionUnsupported(t *testing.T) {
	t.Skip("the overlay compatibility boundary begins at Go 1.19")
}
