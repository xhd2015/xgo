package race_inherit_test

import (
	"sync"
	"testing"

	"github.com/xhd2015/xgo/runtime/tls"
)

var inheritKey = tls.DeclareInherit("race_inherit_key")
var noInheritKey = tls.Declare("race_no_inherit_key")

// TestTLSInheritAcrossGoNoRace is a green canary under -race (no --xgo-race-safe).
//
// What it guards:
//   - TLS inherit into child goroutines must remain correct under -race.
//   - Freelist reuse of *g / __xgo_g must not trip the race detector
//     (exit/create use racerelease/raceacquire on __xgo_g; see #341).
//
// Multi-round nested go churns the g freelist so residual __xgo_g state is
// exercised across owners. This test is expected to PASS; it is not a
// RED-first fixture for clear-vs-release ordering.
//
// If this ever fails with DATA RACE on __xgo_g / G / TLS paths, review:
//  1. __xgo_callback_on_exit_g: racerelease present; clear of __xgo_g should
//     happen-before that release (or only clear after raceacquire on create).
//  2. __xgo_callback_on_create_g: raceacquire then clear residual, then inherit.
//  3. racegostart runs after the create callback (parent inherit writes HB).
//  4. All mirrored trap copies (patches/assets/runtime_gen) stay in sync.
func TestTLSInheritAcrossGoNoRace(t *testing.T) {
	const rounds = 20
	const n = 32
	for r := 0; r < rounds; r++ {
		inheritKey.Set(42 + r)
		noInheritKey.Set(7)
		var wg sync.WaitGroup
		errs := make(chan string, n*2)
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func(want int) {
				defer wg.Done()
				// nested go to churn freelist + inherit chain
				var inner sync.WaitGroup
				inner.Add(1)
				go func() {
					defer inner.Done()
					v, ok := inheritKey.GetOK()
					if !ok {
						errs <- "missing nested inherit"
						return
					}
					if v.(int) != want {
						errs <- "bad nested inherit"
					}
				}()
				v, ok := inheritKey.GetOK()
				if !ok {
					errs <- "missing inherit"
					return
				}
				if v.(int) != want {
					errs <- "bad inherit value"
				}
				if _, ok := noInheritKey.GetOK(); ok {
					errs <- "non-inherit leaked"
				}
				inner.Wait()
			}(42 + r)
		}
		// parent mutates while children still running
		inheritKey.Set(42 + r)
		wg.Wait()
		close(errs)
		for e := range errs {
			t.Error(e)
		}
	}
}
