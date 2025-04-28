package overlay

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/test/build/unified/unified_overlay/overlay/pkg_a"
)

func TestBase(t *testing.T) {
	pkg_a.TestBase(t)
}

func TestPkgA(t *testing.T) {
	pkg_a.TestPkgA(t)
}
