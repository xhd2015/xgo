package shadowed_init

import (
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/mock"
)

// The type of timeout is resolved from its initializer in the scope of
// the file declaring it, not in that of TimeoutSeconds, where newTimeout
// is a local int.
func TestMockVarWithInitializerShadowedAtReference(t *testing.T) {
	mock.Patch(&timeout, func() time.Duration { return 3 * time.Second })
	if got := TimeoutSeconds(); got != 3 {
		t.Fatalf("timeout not mocked: %v", got)
	}
}
