package pkg_b

import (
	"os"
	"testing"
)

func TestBase(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get pwd: %s", err)
	}
	t.Logf("pkg_b.Base: pwd: %s", dir)
}

func TestPkgB(t *testing.T) {
	t.Logf("pkg_b.PkgB")
}
