package std_calls_no_alloc

import (
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/mock"
)

// xgo test lists the predefined std packages (time, os, net/http...) to load
// them, which made them initial packages and trapped every function in them
// instead of only their whitelists. Those compiler-inserted traps are not
// guarded, so calls like time.Time.Sub allocated on every call.
func TestStdCallsOutsideWhitelistDoNotAllocate(t *testing.T) {
	now := time.Now()
	then := now.Add(-3 * time.Hour)
	allocs := testing.AllocsPerRun(100, func() {
		if then.IsZero() {
			t.Fatal("zero")
		}
		_ = now.Sub(then).Hours()
	})
	if allocs != 0 {
		t.Fatalf("time calls allocated %v times per run, want 0", allocs)
	}
}

func TestWhitelistedStdFuncStillPatchable(t *testing.T) {
	fixed := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.Patch(time.Now, func() time.Time {
		return fixed
	})
	if got := time.Now(); !got.Equal(fixed) {
		t.Fatalf("time.Now() = %v, want patched %v", got, fixed)
	}
}
