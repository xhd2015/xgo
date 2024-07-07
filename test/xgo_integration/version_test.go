package xgo_integration

import (
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/cmd"
)

func TestVersion(t *testing.T) {
	checkXgo(t)
	version, err := cmd.Output("xgo", "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(version, "1.0.") {
		t.Fatalf("expect version to be 1.0.x, actual: %s", version)
	}
	suffix := strings.TrimPrefix(version, "1.0.")
	i, err := strconv.ParseInt(suffix, 10, 64)
	if err != nil || i < 0 {
		t.Fatalf("expect version to be 1.0.x, actual: %s", version)
	}
}
