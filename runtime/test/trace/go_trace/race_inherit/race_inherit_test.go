package race_inherit_test

import (
	"sync"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func hello(s string) string {
	return "hello " + s
}

// TestMockInheritAcrossGoNoRace is a green canary under -race (no --xgo-race-safe).
//
// What it guards (issue #341):
//   - Mock state inherited into child goroutines must work under -race.
//   - Create-g inherit must not race with the race detector's go edge
//     (racegostart runs after __xgo_callback_on_create_g).
//   - Freelist reuse of __xgo_g must not trip DATA RACE (exit/create
//     racerelease/raceacquire; multi-round nested go churns freelist).
//
// Expected PASS. This is a regression guard, not a RED-first clear-order test.
//
// If this ever fails with DATA RACE on __xgo_g / mock inherit paths, review:
//  1. racegostart after __xgo_callback_on_create_g (parent→child inherit HB).
//  2. __xgo_callback_on_exit_g: racerelease; prefer clear of __xgo_g before release.
//  3. __xgo_callback_on_create_g: raceacquire then clear residual, then inherit.
//  4. Mirrored trap copies (patches/assets/runtime_gen) stay in sync.
func TestMockInheritAcrossGoNoRace(t *testing.T) {
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
