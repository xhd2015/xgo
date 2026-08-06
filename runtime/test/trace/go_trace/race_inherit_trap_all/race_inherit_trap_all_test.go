package race_inherit_trap_all_test

import (
	"sync"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func hello(s string) string {
	return "hello " + s
}

// TestMockInheritAcrossGoNoRace_trap_all is a green canary under
// -race --trap-all (no --xgo-race-safe).
//
// Same inherit / freelist guards as race_inherit, plus trap-all re-entrancy
// during create-g inherit (e.g. time methods used for beginNs deltas).
//
// Expected PASS. If DATA RACE or hang appears, review create-g racegostart
// order, exit/create racerelease/raceacquire, and trap-all re-entry on inherit.
func TestMockInheritAcrossGoNoRace_trap_all(t *testing.T) {
	mock.Patch(hello, func(s string) string {
		return "mocked " + s
	})

	const rounds = 20
	const n = 32
	for r := 0; r < rounds; r++ {
		var wg sync.WaitGroup
		errs := make(chan string, n*2)
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				// nested go: freelist churn + inherit chain (same shape as TLS canary)
				var inner sync.WaitGroup
				inner.Add(1)
				go func() {
					defer inner.Done()
					got := hello("x")
					if got != "mocked x" {
						errs <- "nested: " + got
					}
				}()
				got := hello("x")
				if got != "mocked x" {
					errs <- got
				}
				inner.Wait()
			}()
		}
		wg.Wait()
		close(errs)
		for got := range errs {
			t.Errorf("expect mock inherited into child goroutine, got %q", got)
		}
	}
}
