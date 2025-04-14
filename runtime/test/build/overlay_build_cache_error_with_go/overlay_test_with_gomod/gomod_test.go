package overlay_build_cache_error

import (
	"strings"
	"testing"

	"golang.org/x/example/hello/reverse"
)

func TestReverse(t *testing.T) {
	res := reverse.String("hello")
	if res != "olleh" && !strings.HasPrefix(res, "ollehgo") {
		t.Fatalf("res=%s", res)
	}
}
