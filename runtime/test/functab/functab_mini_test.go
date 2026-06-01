package functab

import (
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/functab"
)

func Hello(s string) string {
	return "hello " + s
}

func TestFunctabMini(t *testing.T) {
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
