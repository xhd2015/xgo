# API Summary
When passing a variable pointer to `Patch` and `Mock`, xgo will lookup for package level variables and setup trap for accessing to these variables.

Constant can only be patched via `PatchByName(pkg,name,replacer)`.

# Limitation
1. Only variables and consts of main module will be available for patching,
2. Constant patching requires go>=1.20.

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

Check [../test/patch_const/patch_var_test.go](../test/patch_const/patch_var_test.go) for more cases.

## `PatchByName` on constant
```go
package patch_const

import (
    "testing"
    
    "github.com/xhd2015/xgo/runtime/mock"
)

const N = 50

func TestPatchInElseShouldWork(t *testing.T) {
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