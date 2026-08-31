# API Summary
When passing a variable pointer to `Patch` and `Mock`, xgo will lookup for package level variables and setup trap for accessing to these variables.

Constant can only be patched via `PatchByName(pkg,name,replacer)`.

# Limitation
1. Only variables and consts of main module will be available for patching,
2. Constant patching requires go>=1.20.

# Value Patch vs `&var` (no VarPtr fallback)

Instrumented reads are split into two channels:

| Access | Rewrite | Trap | Satisfied by |
|---|---|---|---|
| `x := pkg.Var` | `pkg.Var_xgo_get()` | Var (`trapVar`) | `mock.Patch(&pkg.Var, func() T { ... })` |
| `p := &pkg.Var` | `pkg.Var_xgo_get_addr()` | VarPtr (`trapVarPtr`) | `mock.Patch(&pkg.Var, func() *T { ... })` |

A **value** Patch (`func() T`) does **not** apply to `&pkg.Var`.

**Why:** Var→VarPtr fallback used to apply the value mock when taking the address of a package var. That path could **overwrite package storage** and leak across tests, so it is disabled (`DISABLE_PTR_FALLBACK` in `runtime/internal/trap/var.go`). Regression: `TestPatchVarPtrShouldNotFallbackTest` in [../test/patch/patch_var_test.go](../test/patch/patch_var_test.go).

**Workarounds** when production code passes `&pkg.Var` into an API that reads through the pointer (e.g. `Hit(ctx, &cfg, id)`):

1. **Copy, then take address** (uses value read / Var trap):

```go
v := pkg.Var
use(&v)
```

2. **Explicit VarPtr Patch** (replacer must return `*T` to a stable value, not a dangling stack temp):

```go
mock.Patch(&pkg.Var, func() *T {
    v := want
    return &v // ok if only read during the same call; prefer a heap/stable pointer for longer use
})
```

# Examples
## `Patch` on variable
```go
package patch

import (
    "testing"

    "github.com/xhd2015/xgo/runtime/mock"
)

var a int = 123

func TestPatchVarTest(t *testing.T) {
	mock.Patch(&a, func() int {
		return 456
	})
	b := a
	if b != 456 {
		t.Fatalf("expect patched variable a to be %d, actual: %d", 456, b)
	}
}

```

Check [../test/patch/patch_var_test.go](../test/patch/patch_var_test.go) for more cases.

## `PatchByName` on constant
```go
package patch_const

import (
    "testing"
    
    "github.com/xhd2015/xgo/runtime/mock"
)

const N = 50

func TestPatchConst(t *testing.T) {
    mock.PatchByName("github.com/xhd2015/xgo/runtime/test/patch_const", "N", func() int {
        return 5
    })
    b := N*4

    if b != 20 {
        t.Fatalf("expect b to be %d,actual: %d", 20, b)
    }
}
```

Check [../test/patch_const/patch_const_test.go](../test/patch_const/patch_const_test.go) for more cases.