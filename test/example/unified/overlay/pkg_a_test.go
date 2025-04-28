package overlay

import (
	"testing"

	"github.com/xhd2015/xgo/test/example/unified/overlay/pkg_a"
)

func TestBase(t *testing.T) {
	pkg_a.TestBase(t)
}

func TestPkgA(t *testing.T) {
	pkg_a.TestPkgA(t)
}
