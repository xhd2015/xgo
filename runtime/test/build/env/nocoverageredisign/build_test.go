//go:build go1.20
// +build go1.20

package nocoverageredisign

import (
	"os"
	"testing"
)

func TestNoCoverageRedesign(t *testing.T) {
	v := os.Getenv("GOEXPERIMENT")
	t.Logf("GOEXPERIMENT: %s", v)
	if v != "nocoverageredesign" {
		t.Fatalf("expected GOEXPERIMENT to be nocoverageredesign, but got %s", v)
	}
}
