package race_inherit_trap_stdlib_test

import (
	"sync"
	"testing"

	"github.com/xhd2015/xgo/runtime/tls"
)

var inheritKey = tls.DeclareInherit("race_inherit_key")
var noInheritKey = tls.Declare("race_no_inherit_key")

// TestTLSInheritAcrossGoNoRace_trap_stdlib is a green canary under
// -race --trap-stdlib (no --xgo-race-safe).
//
// Same TLS inherit / freelist guards as race_inherit, plus stdlib trap
// pressure during create-g inherit.
//
// Expected PASS. If DATA RACE or hang appears, review exit/create
// racerelease/raceacquire and racegostart-after-callback ordering.
func TestTLSInheritAcrossGoNoRace_trap_stdlib(t *testing.T) {
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
