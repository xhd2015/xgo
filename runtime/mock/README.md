# API Summary
> This document describes `Mock` and `Patch` for functions. For variables and consts, see [MOCK_VAR_CONST.md](./MOCK_VAR_CONST.md).

Mock exposes 3 `Mock` APIs to users:
- `Mock(fn, interceptor)` - for **99%** scenarios

- `MockByName(pkgPath, name, interceptor)` - for **unexported** function

- `MockMethodByName(instance, name, interceptor)` - for **unexported** method

And another 3 `Patch` APIs:
- `Patch(fn, replacer)` - for **99%** scenarios

- `PatchByName(pkgPath, name, replacer)` - for **unexported** function

- `PatchMethodByName(instance, name, replacer)` - for **unexported** method

Under 99% circumstances, developer should use `Mock` or `Patch` as long as possible because it does not involve hard coded name or package path.

The later two, `MockByName` and `MockMethodByName` are used where the target method cannot be accessed due to unexported, so they must be referenced by hard coded strings.

All 3 mock APIs take in an `InterceptorFunc` as the last argument,  and returns a `func()`. The returned `func()` can be used to clear the passed in interceptor.

# Difference between `Mock` and `Patch`
The difference lies at their last parameter:
- `Mock` accepts an `InterceptorFunc`
- `Patch` accepts a function with the same signature as the first parameter

# Scope
Based on the timing when `Mock*`,or `Patch*` is called, the interceptor has different behaviors:
- If called from `init`, then all goroutines will be mocked,
- Otherwise, `Mock*` or `Patch*` is called after `init`, then the mock interceptor will only be effective for current gorotuine, other goroutines are not affected.

# Interceptor
Signature: `type InterceptorFunc func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error`

- If the interceptor returns `nil`, then the target function is mocked,
- If the interceptor returns `mock.ErrCallOld`, then the target function is called again,
- Otherwise, the interceptor returns a non-nil error, that will be set to the function's return error.

# Mock
Signature: `Mock(fn interface{}, interceptor InterceptorFunc) func()`

Example:
```go
package demo

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func MyFunc() string {
	return "my func"
}
func TestFuncMock(t *testing.T) {
	mock.Mock(MyFunc, func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error {
		results.GetFieldIndex(0).Set("mock func")
		return nil
	})
	text := MyFunc()
	if text != "mock func" {
		t.Fatalf("expect MyFunc() to be 'mock func', actual: %s", text)
	}
}
```

# MockByName
Signature:  `MockByName(pkgPath string, name string, interceptor InterceptorFunc) func()`

Example:
```go
package demo

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"

	"github.com/my/another_pkg"
)

func TestGetNameShouldFail(t *testing.T) {
	mock.MockByName("github.com/my/another_pkg", "getNameThroughHTTP", func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
        results.FieldIndex(0).Set("failed")
		return nil
	})
    name := another_pkg.GetName()
    if name != "failed"{
        t.Fatalf("expect GetName()=%s, actual: %s","failed",name)
    }
}
```

Another package:
```go
package another_pkg

func GetName() string {
    ...
    if notHitCache {
        name = getNameThroughHTTP()
        setCache(name)
    }
    return name
}

func getNameThroughHTTP() string {
    ...
    http.Do(....)
    ...
    return "ok"
}
```

# MockMethodByName
Signature: `MockMethodByName(instance interface{}, name string, interceptor InterceptorFunc) func()`

Example:
```go

package demo

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"

	"github.com/my/another_pkg"
)

func TestGetNameShouldFail(t *testing.T) {
    client := another_pkg.GetClient()

	mock.MockMethodByName(client, "getNameThroughHTTP",func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
        results.FieldIndex(0).Set("failed")
		return nil
	})
    name := client.GetName()
    if name != "failed"{
        t.Fatalf("expect GetName()=%s, actual: %s","failed",name)
    }
}
```

Another package:
```go
package another_pkg

type client struct {
    endpoint string
}

type Client interface {
    GetName() string
}

var defaultClient = &client{endpoint:"default"}

func GetClient() Client {
    return defaultClient
}

func (c *client) GetName() string {
    ...
    if notHitCache {
        name = c.getNameThroughHTTP()
        setCache(name)
    }
    return name
}

func (c *client) getNameThroughHTTP() string {
    ...
    http.Do(....)
    ...
    return "ok"
}
```

# Mock Generic Function
In the following functions, `ToString[int]` gets mocked, while `ToString[string]` does not.
```go
package mock_generic

import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func ToString[T any](v T) string {
	return fmt.Sprint(v)
}

func TestMockGenericFunc(t *testing.T) {
	mock.Mock(ToString[int], func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set("hello world")
		return nil
	})
	expect := "hello world"
	output := ToString[int](0)
	if expect != output {
		t.Fatalf("expect ToString[int](0) to be %s, actual: %s", expect, output)
	}

	// per t param
	expectStr := "my string"
	outputStr := ToString[string](expectStr)
	if expectStr != outputStr {
		t.Fatalf("expect ToString[string](%q) not affected, actual: %q", expectStr, outputStr)
	}
}
```

# Patch
```go
package patch_test

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func greet(s string) string {
	return "hello " + s
}

func TestPatchFunc(t *testing.T) {
	mock.Patch(greet, func(s string) string {
		return "mock " + s
	})

	res := greet("world")
	if res != "mock world" {
		t.Fatalf("expect patched result to be %q, actual: %q", "mock world", res)
	}
}
```