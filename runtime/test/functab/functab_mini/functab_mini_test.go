package functab_mini

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trap"
)

func Hello(s string) string {
	return "hello " + s
}

func TestFunctabMini(t *testing.T) {
	// Theory C verification: if overlay is applied, trap interceptor will fire
	var intercepted bool
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			if f.IdentityName == "Hello" {
				intercepted = true
			}
			return nil, nil
		},
	})
	_ = Hello("test")
	if intercepted {
		t.Logf("TRAP_CHECK: interceptor fired for Hello — overlay IS applied")
	} else {
		t.Errorf("TRAP_CHECK: interceptor did NOT fire for Hello — overlay may NOT be applied (source compiled without __xgo_trap_0)")
	}

	funcs := functab.GetFuncs()

	t.Logf("functab.GetFuncs() returned %d entries", len(funcs))
	for _, fi := range funcs {
		t.Logf("  - %s (kind=%v, identity=%s)", fi.FullName, fi.Kind, fi.IdentityName)
	}

	found := false
	for _, fi := range funcs {
		if fi.IdentityName == "Hello" {
			found = true
			t.Logf("Found Hello in functab: fullName=%s kind=%v", fi.FullName, fi.Kind)
			break
		}
	}

	if !found {
		// Try InfoFunc as well
		info := functab.InfoFunc(Hello)
		if info != nil {
			t.Logf("InfoFunc(Hello) returned: identity=%s, fullName=%s", info.IdentityName, info.FullName)
		} else {
			t.Logf("InfoFunc(Hello) returned nil")
		}
		t.Errorf("Hello not found in functab. Got %d funcs, need at least 1 (Hello). Check if xgo instrumentation is working.", len(funcs))
	}

	// Verify that at minimum we got the Hello function registered
	if len(funcs) == 0 {
		t.Errorf("functab is completely empty (0 entries). This means xgo link rewriting did not occur. Enable XGO_DEBUG_LINK_INIT=true to diagnose.")
	}

	fmt.Printf("DEBUG: functab has %d entries\n", len(funcs))
}
