package amd64

import (
	"os"
	"testing"
)

func TestNoCoverageRedesign(t *testing.T) {
	v := os.Getenv("GOARCH")
	t.Logf("GOARCH: %s", v)
	if v != "amd64" {
		t.Fatalf("expected GOARCH to be amd64, but got %s", v)
	}
}
