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
	minimumMappingUID = userNsLength
	mappingLen        = userNsLength * 2000
)

// see bug https://github.com/xhd2015/xgo/issues/176
func TestUntypedUnknownConstShouldCompile(t *testing.T) {
	var allocated uint32

	if allocated > minimumMappingUID+mappingLen-userNsLength {
		t.Fatalf("allocated is greater?")
	}
}
