package windows

import (
	"os"
	"testing"
)

func TestNoCoverageRedesign(t *testing.T) {
	v := os.Getenv("GOOS")
	t.Logf("GOOS: %s", v)
	if v != "windows" {
		t.Fatalf("expected GOOS to be windows, but got %s", v)
	}
}
