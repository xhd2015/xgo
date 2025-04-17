package arm64

import (
	"os"
	"testing"
)

func TestNoCoverageRedesign(t *testing.T) {
	v := os.Getenv("GOARCH")
	t.Logf("GOARCH: %s", v)
	if v != "arm64" {
		t.Fatalf("expected GOARCH to be arm64, but got %s", v)
	}
}
