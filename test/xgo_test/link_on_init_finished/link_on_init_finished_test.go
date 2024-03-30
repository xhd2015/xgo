package trap_set

import (
	"fmt"
	"os"
	"testing"
)

func __xgo_link_on_init_finished(f func()) {
	fmt.Fprintln(os.Stderr, "WARNING: failed to link __xgo_link_on_init_finished.(xgo required)")
}

var ran bool

func init() {
	__xgo_link_on_init_finished(runAfterInit)
}

// go run ./cmd/xgo test -v -run TestLinkOnFinished ./test/xgo_test/link_on_init_finished
func TestLinkOnInitFinished(t *testing.T) {
	if !ran {
		t.Fatalf("expect have called runAfterInit, actually not called")
	}
}

func runAfterInit() {
	ran = true
}
