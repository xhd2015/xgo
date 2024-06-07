// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code
//
// usage:
//  go run -tags dev ./cmd/xgo test --project-dir runtime/test/debug
//  go run -tags dev ./cmd/xgo test --debug-compile --project-dir runtime/test/debug

package debug

import (
	"testing"
)

// see bug https://github.com/xhd2015/xgo/issues/176
func TestNodef(t *testing.T) {
}
