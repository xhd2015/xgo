package external_overlay_composition

// This import is deliberately absent. Go 1.24+ vet must read the caller
// overlay replacement instead of this original source file.
import _ "github.com/xhd2015/xgo/runtime/test/build/external_overlay_composition/not_present"

func Value() string {
	return "original"
}
