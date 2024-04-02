// when multiple interceptors added, the order is reversed
package trap

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestTrapOverlay(t *testing.T) {
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "testOverlay" {
				panic("first trap should not be called")
			}
			return nil, nil
		},
	})

	// overlay
	var trapCalled bool
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "testOverlay" {
				trapCalled = true
				return nil, trap.ErrAbort
			}
			return
		},
	})
	testOverlay()

	if !trapCalled {
		t.Fatalf("expect trap to have been called, actually not called")
	}
}

func testOverlay() {

}
