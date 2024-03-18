# xgo
Enable function Trap for `go`, and provide tools like Trace and Mock to help go developers write unit test and debug both easier and faster.

`xgo` works as a drop-in replacement for `go run`,`go build`, and `go test`.

It adds missing abilities to go program by cooperating with(or hacking) the go compiler.

Thses abilities include [Mock](#mock), [Trap](#trap) and [Trace](#trace).

See [Usage](#usage) and [Documentation](#xgo) for more details. 


# Install
```sh
# macOS and Linux
curl -fsSL https://github.com/xhd2015/xgo/raw/master/install.sh | bash
```

# Usage
The following code shows how to setup mock on a given function:
- The function `add(a,b)` normally adds `a` and `b` up, resulting in `a+b`,
- However, after `mock.AddFuncInterceptor`, the logic is changed to `a-b`

(check [test/testdata/mock_res/main.go](test/testdata/mock_res/main.go) for more details.)
```go
package main

import (
	"context"
	"fmt"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
)

func main() {
	before := add(5, 2)
	fmt.Printf("before mock: add(5,2)=%d\n", before)
    mock.AddFuncInterceptor(add, func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error {
        a := args.GetField("a").Value().(int)
        b := args.GetField("b").Value().(int)
        res := a - b
        results.GetFieldIndex(0).Set(res)
        return nil
    })
	after := add(5, 2)
	fmt.Printf("after mock: add(5,2)=%d\n", after)
}

func add(a int, b int) int {
	return a + b
}

```

Run with `go`:

```sh
go run ./
# output:
#  before mock: add(5,2)=7
#  after mock: add(5,2)=7
```

Run with `xgo`:

```sh
xgo run ./
# output:
#  before mock: add(5,2)=7
#  after mock: add(5,2)=3
```

# Trap
Trap allows developer to intercept function execution on the fly.

Trap is the core of `xgo` as it is the basis of other abilities like Mock and Trace.

The following example logs function execution trace by adding a Trap interceptor:

(check [test/testdata/trap/trap.go](test/testdata/trap/trap.go) for more details.)
```go
package main

import (
	"context"
	"fmt"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func init() {
    trap.AddInterceptor(&trap.Interceptor{
        Pre: func(ctx context.Context, f *core.FuncInfo, args core.Object, results core.Object) (interface{}, error) {
            trap.Skip()
            if f.Name == "A" {
                fmt.Printf("trap A\n")
                return nil, nil
            }
            if f.Name == "B" {
                fmt.Printf("abort B\n")
                return nil, trap.ErrAbort
            }
            return nil, nil
        },
    })
}

func main() {
	A()
	B()
}

func A() {
	fmt.Printf("A\n")
}

func B() {
	fmt.Printf("B\n")
}
```

Run with `go`:

```sh
go run ./
# output:
#   A
#   B
```

Run with `xgo`:

```sh
xgo run ./
# output:
#   trap A
#   A
#   abort B
```

Trap has two APIs to add function interceptor:
- `AddInterceptor()`:  effective globally for all goroutines,
- `AddLocalInterceptor()`: effective only for current goroutine.

# Mock
Mock simplifies the process of setting up Trap interceptors.

It exposes 1 API:
- `AddFuncInterceptor()`

The detailed usage can be found in [Usage](#usage) section.

# Trace
It is painful when debugging with a deep call stack.

Trace addresses this issue by collecting the hiearchical stack trace and stores it into file for later use. 

Needless to say, with Trace, debug becomes less usual:

(check [test/testdata/trace/trace.go](test/testdata/trace/trace.go) for more details.)
```go
package main

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/trace"
)

func init() {
	trace.Use()
}

func main() {
	A()
	B()
	C()
}
func A() { fmt.Printf("A\n") }
func B() { fmt.Printf("B\n");C(); }
func C() { fmt.Printf("C\n") }
```

Run with `go`:

```sh
go run ./
# output:
#   A
#   B
#   C
#   C
```

Run with `xgo`:

```sh
XGO_TRACE_OUTPUT=stdout xgo run ./
# output a JSON representing the call stacktrace like:
#        {
#            "Name": "main",
#            "Children": [{
#                 "Name": "A"
#                },{
#                  "Name": "B",
#                  "Children": [{
#                       "Name": "C"
#                   }]
#                },{
#                   "Name": "C"
#                }
#            ]
#        }
#
# NOTE: other fields are ommited for displaying key information.
```

By default, Trace write traces to a temp directory under current working directory.This behavior can be overridden by setting `XGO_TRACE_OUTPUT` to different values:
- `XGO_TRACE_OUTPUT=stdout`: traces will be written to stdout, for debugging purepose,
- `XGO_TRACE_OUTPUT=<dir>`: traces will be written to `<dir>`,
- `XGO_TRACE_OUTPUT=off`: turn off trace.

# Evolution of `xgo`
`xgo` is the successor of the original [go-mock](https://github.com/xhd2015/go-mock), which works by rewriting go code before compile.

That strategy works well but causes much longer build time for larger projects due to source code explosion.

However, `go-mock` is remarkable for it's discovery of Trap, Trace besides Mock, and additional abilities like trapping variable and disabling map randomness.

It is the shoulder which `xgo` stands on.

# Go Version Requirement
`xgo` requires at least `go1.17` to compile.
