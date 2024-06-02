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

const (
	pod1 = "pod1"
)

type Pod struct {
	Name string
}

func TestConstNameCollision(t *testing.T) {
	var pod1 *Pod

	if pod1 != nil && pod1.Name != "" {
		t.Fatalf("pod1 should be empty")
	}
}
