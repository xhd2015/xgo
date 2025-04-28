package pkg_a

import (
	"os"
	"testing"

	"github.com/xhd2015/xgo/runtime/test/build/unified/shared"
)

func init() {
	shared.IncrementTest()
}

func TestBase(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get pwd: %s", err)
	}
	t.Logf("pkg_a.Base: pwd: %s", dir)
}

func TestPkgA(t *testing.T) {
	t.Logf("pkg_a.PkgA")
}
