package shared

import (
	"testing"
)

// see https://github.com/xhd2015/xgo/issues/350
func TestShared(t *testing.T) {
	if numTest != 2 {
		t.Errorf("expect numTest to be 2, actual: %d", numTest)
	}
}
