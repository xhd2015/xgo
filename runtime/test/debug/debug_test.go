// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code
//
// usage:
//     go run -tags dev ./cmd/xgo test --debug-compile --project-dir runtime/test/debug

package debug

import (
	"testing"
)

const userNsLength = (1 << 16)
const (
	// skip the first block
	minimumMappingUID = userNsLength
	// allocate enough space for 2000 user namespaces
	mappingLen = userNsLength * 2000
)

func TestArrayPointer(t *testing.T) {
	var allocated uint32

	if allocated <= minimumMappingUID+mappingLen-userNsLength {

	}
}
