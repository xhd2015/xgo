package overlay

import (
	"testing"

	"github.com/xhd2015/xgo/test/example/unified/overlay/pkg_b"
)

func TestBaseX(t *testing.T) {
	pkg_b.TestBase(t)
}

func TestPkgB(t *testing.T) {
	pkg_b.TestPkgB(t)
}
