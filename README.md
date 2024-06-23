
# xgo

[![Go Reference](https://pkg.go.dev/badge/github.com/xhd2015/xgo.svg)](https://pkg.go.dev/github.com/xhd2015/xgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/xhd2015/xgo)](https://goreportcard.com/report/github.com/xhd2015/xgo)
[![Slack Widget](https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=red)](https://join.slack.com/t/golang-xgo/shared_invite/zt-2ixe161jb-3XYTDn37U1ZHZJSgnQi6sg)
[![Go Coverage](https://img.shields.io/badge/Coverage-82.6%25-brightgreen)](https://github.com/xhd2015/xgo/actions)
[![CI](https://github.com/xhd2015/xgo/workflows/Go/badge.svg)](https://github.com/xhd2015/xgo/actions)
[![Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

**English | [简体中文](./README_zh_cn.md)**

`xgo` provides *all-in-one* test utilities for golang, including:

- [API](#api):
  - [Mock](#patch)
  - [Trace](#trace)
  - [Trap](#trap)
- [Tools](#tools):
  - [Test Explorer](#test-explorer)
  - [Incremental Coverage](#incremental-coverage)

As for the monkey patching part, `xgo` works as a preprocessor for `go run`,`go build`, and `go test`(see our [blog](https://blog.xhd2015.xyz/posts/xgo-monkey-patching-in-go-using-toolexec)).

See [Quick Start](#quick-start) and [Documentation](./doc) for more details.

[![Xgo Test Explorer](https://github.com/xhd2015/xgo/assets/14964938/53de1de1-85a6-41b0-9300-c4979ba0096c)](https://youtu.be/L8oPhgqtdGg)

![Xgo Test Explorer](https://github.com/xhd2015/xgo/assets/14964938/2666f0c9-3a9f-45e0-a1e2-20e1bf401fa1)

> *By the way, I promise you this is an interesting project.*

# Installation
```sh
go install github.com/xhd2015/xgo/cmd/xgo@latest
```

Verify the installation:
```sh
xgo version
# output:
#   1.0.x

xgo help
# output: help messages
```

If `xgo` is not found, you may need to check if `$GOPATH/bin` is added to your `PATH` variable.

For CI jobs like github workflow, see [doc/INSTALLATION.md](./doc/INSTALLATION.md).

# Requirement
`xgo` requires at least `go1.17` to compile.

There is no specific limitation on OS and Architecture. 

**All OS and Architectures** are supported by `xgo` as long as they are supported by `go`.
|         | x86 | x86_64 (amd64)  | arm64     | any other Arch...     |
|:---------|:-----------:|:-----------:|:-----------:|:-----------:|
| Linux   | Y | Y | Y | Y |
| Windows | Y | Y | Y | Y |
| macOS   | Y | Y | Y | Y |
| any other OS... | Y | Y | Y | Y|

# Quick Start
Let's write a unit test with `xgo`:

1. Ensure you have installed `xgo` by following the [Installation](#installation) section, and verify the installation with:
```sh
xgo version
# output
#   1.0.x
```

2. Init a go project:
```sh
mkdir demo
cd demo
go mod init demo
```
3. Add `demo_test.go`:
```go
package demo_test

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func MyFunc() string {
	return "my func"
}

func TestFuncMock(t *testing.T) {
	mock.Patch(MyFunc, func() string {
		return "mock func"
	})
	text := MyFunc()
	if text != "mock func" {
		t.Fatalf("expect MyFunc() to be 'mock func', actual: %s", text)
	}
}
```
4. Get the `github.com/xhd2015/xgo/runtime` dependency:
```sh
go get github.com/xhd2015/xgo/runtime
```
5. Test the code:
```sh
xgo test -v ./
# NOTE: xgo will take some time to setup for the first time
```

Output:
```sh
=== RUN   TestFuncMock
--- PASS: TestFuncMock (0.00s)
PASS
ok      demo
```

NOTE: `xgo` is used to test your code, not just `go`. 

Under the hood, `xgo` preprocess your code to add mock hooks, and then calls `go` to do remaining jobs.

The above code can be found at [doc/demo](./doc/demo).

# API

## Patch
Patch given function within current goroutine.

The API:
- `Patch(fn,replacer) func()`

Cheatsheet:
```go
// package level func
mock.Patch(SomeFunc, mockFunc)

// per-instance method
// only the bound instance `v` will be mocked
// `v` can be either a struct or an interface
mock.Patch(v.Method, mockMethod)

// per-TParam generic function
// only the specified `int` version will be mocked
mock.Patch(GenericFunc[int], mockFuncInt)

// per TParam and instance generic method
v := GenericStruct[int]
mock.Mock(v.Method, mockMethod)

// closure can also be mocked
// less used, but also supported
mock.Mock(closure, mockFunc)
```

Parameters:
- If `fn` is a simple function(i.e. a package level function, or a function owned by a type, or a closure(yes, we do support mocking closures)),then all call to that function will be intercepted,
- `replacer` another function that will replace `fn`

Scope:
- If `Patch` is called from `init`, then all goroutines will be mocked.
- Otherwise, `Patch` is called after `init`, then the mock interceptor will only be effective for current goroutine, other goroutines are not affected.

NOTE: `fn` and `replacer` should have the same signature.

Return:
- a `func()` can be used to remove the replacer earlier before current goroutine exits

Patch replaces the given `fn` with `replacer` in current goroutine. It will remove the replacer once current goroutine exits.

Example:
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

NOTE: `Patch` and `Mock`(below) supports top-level variables and consts, see [runtime/mock/MOCK_VAR_CONST.md](runtime/mock/MOCK_VAR_CONST.md).

**Notice for mocking stdlib**: There are different modes for mocking stdlib functions,see [runtime/mock/stdlib.md](./runtime/mock/stdlib.md).

## Mock
`runtime/mock` also provides another API called `Mock`, which is similar to `Patch`.

The the only difference between them lies in the the second parameter: `Mock` accepts an interceptor. 

`Mock` can be used where `Patch` cannot be used, such as functions with unexported type.

> API details: [runtime/mock/README.md](runtime/mock)

The Mock API:
- `Mock(fn, interceptor)`

Parameters:
- `fn` same as described in [Patch](#patch) section
- If `fn` is a method(i.e. `file.Read`),then only call to the instance will be intercepted, other instances will not be affected

Interceptor Signature: `func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error`
- If the interceptor returns `nil`, then the target function is mocked,
- If the interceptor returns `mock.ErrCallOld`, then the target function is called again,
- Otherwise, the interceptor returns a non-nil error, that will be set to the function's return error.

There are other 2 APIs can be used to setup mock based on name, check [runtime/mock/README.md](runtime/mock/README.md) for more details.

Method mock example:
```go
type MyStruct struct {
    name string
}
func (c *MyStruct) Name() string {
    return c.name
}

func TestMethodMock(t *testing.T){
    myStruct := &MyStruct{
        name: "my struct",
    }
    otherStruct := &MyStruct{
        name: "other struct",
    }
    mock.Mock(myStruct.Name, func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error {
        results.GetFieldIndex(0).Set("mock struct")
        return nil
    })

    // myStruct is affected
    name := myStruct.Name()
    if name!="mock struct"{
        t.Fatalf("expect myStruct.Name() to be 'mock struct', actual: %s", name)
    }

    // otherStruct is not affected
    otherName := otherStruct.Name()
    if otherName!="other struct"{
        t.Fatalf("expect otherStruct.Name() to be 'other struct', actual: %s", otherName)
    }
}
```

## Trace
> Trace might be the most powerful tool provided by xgo, this blog has a more thorough example: https://blog.xhd2015.xyz/posts/xgo-trace_a-powerful-visualization-tool-in-go

It is painful when debugging with a deep call stack.

Trace addresses this issue by collecting the hierarchical stack trace and stores it into file for later use.

Needless to say, with Trace, debug becomes less usual:

```go
package trace_test

import (
    "fmt"
    "testing"
)

func TestTrace(t *testing.T) {
    A()
    B()
    C()
}

func A() { fmt.Printf("A\n") }
func B() { fmt.Printf("B\n");C(); }
func C() { fmt.Printf("C\n") }
```

Run with `xgo`:

```sh
# run the test
# this will write the trace into TestTrace.json
# --strace represents stack trace
xgo test --strace ./

# view the trace
xgo tool trace TestTrace.json
```

Output:
![trace html](doc/img/trace_demo.jpg "Simple Trace")

Another more complicated example from [runtime/test/stack_trace/update_test.go](runtime/test/stack_trace/update_test.go): 
![trace html](cmd/xgo/trace/testdata/stack_trace.jpg "Complicatd Trace")

Real world examples: 
- https://github.com/Shibbaz/GOEventBus/pull/11

Trace helps you get started to a new project quickly.

By default, Trace will write traces to a temp directory under current working directory. This behavior can be overridden by setting `XGO_TRACE_OUTPUT` to different values:
- `XGO_TRACE_OUTPUT=stdout`: traces will be written to stdout, for debugging purpose,
- `XGO_TRACE_OUTPUT=<dir>`: traces will be written to `<dir>`,
- `XGO_TRACE_OUTPUT=off`: turn off trace.

Besides the `--strace` flag, xgo allows you to define which span should be collected, using `trace.Begin()`:
```go
import "github.com/xhd2015/xgo/runtime/trace"

func TestTrace(t *testing.T) {
    A()
    finish := trace.Begin()
    defer finish()
    B()
    C()
}
```
The trace will only include `B()` and `C()`.
## Trap
Xgo **preprocess** the source code and IR(Intermediate Representation) before invoking `go`, providing a chance for user to intercept any function when called.

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

`AddInterceptor()` add given interceptor to either global or local, depending on whether it is called from `init` or after `init`:
- Before `init`: effective globally for all goroutines,
- After `init`: effective only for current goroutine, and will be cleared after current goroutine exits.

When `AddInterceptor()` is called after `init`, it will return a dispose function to clear the interceptor earlier before current goroutine exits.

Example:

```go
func main(){
    clear := trap.AddInterceptor(...)
    defer clear()
    ...
}
```

Trap also have a helper function called `Direct(fn)`, which can be used to bypass any trap and mock interceptors, calling directly into the original function.

# Tools
## Test Explorer

The `xgo e` sub command will open a test explorer UI in the browser, provide go developers an easy way to test and debug go code.

With the test explorer, `xgo test` is used instead of `go test`, to enable mocking functionalities.

```sh
$ xgo e
Server listen at http://localhost:7070
```

<img width="1792" alt="test-explorer" src="https://github.com/xhd2015/xgo/assets/14964938/f8689a2d-b422-4101-acef-0aa41ad7e246">

P.S.: `xgo e` is an alias for `xgo tool test-explorer`.

For help, run `xgo e help`.

# Incremental Coverage
The `xgo tool coverage` sub command extends go's builtin `go tool cover` for better visualization.

First, run `go test` or `xgo test` to get a coverage profile:
```sh
go test -cover -coverpkg ./... -coverprofile cover.out ./...
``` 

Then, use `xgo` to display the coverage:
```sh
xgo tool coverage serve cover.out
```

Output:

![coverage](doc/img/coverage.jpg "Coverage")

The displayed coverage is a combination of coverage and git diff. By default, only modified lines were shown:
- Covered lines shown as light blue,
- Uncovered lines shown as light yellow

This helps to quickly locate changes that were not covered, and add tests for them incrementally.

# Concurrent safety
I know you guys from other monkey patching library suffer from the unsafety implied by these frameworks.

But I guarantee you mocking in xgo is builtin concurrent safe. That means, you can run multiple tests concurrently as long as you like.

Why? when you run a test, you setup some mock, these mocks will only affect the goroutine test running the test. And these mocks get cleared when the goroutine ends, no matter the test passed or failed.

Want to know why? Stay tuned, we are working on internal documentation.

# Implementation Details
> Working in progress...

See [Issue#7](https://github.com/xhd2015/xgo/issues/7) for more details.

This blog has a basic explanation: https://blog.xhd2015.xyz/posts/xgo-monkey-patching-in-go-using-toolexec

# Why `xgo`?
The reason is simple: **NO** interface.

Yes, no interface, just for mocking. If the only reason to abstract an interface is to mock, then it only makes me feel boring, not working.

Extracting interface just for mocking is never an option to me. To the domain of the problem, it's merely a workaround. It enforces the code to be written in one style, that's why we don't like it.

Monkey patching simply does the right thing for the problem. But existing library are bad at compatibility.

So I created `xgo`, and hope it will finally take over other solutions to the mocking problem.

# Comparing `xgo` with `monkey`
The project [bouk/monkey](https://github.com/bouk/monkey), was initially created by bouk, as described in his blog https://bou.ke/blog/monkey-patching-in-go.

In short, it uses a low level assembly hack to replace function at runtime. Which exposes lots of confusing problems to its users as it gets used more and more widely(especially on macOS).

Then it was archived and no longer maintained by the author himself. However, two projects later take over the ASM idea and add support for newer go versions and architectures like Apple M1.

Still, the two does not solve the underlying compatibility issues introduced by ASM, including cross-platform support, the need to write to a read-only section of the execution code and lacking of general mock.

So developers still get annoying failures every now and then.

Xgo managed to solve these problems by avoiding low level hacking of the language itself. Instead, it relies on the IR representation employed by the go compiler.

It does a so-called `IR Rewriting` on the fly when the compiler compiles the source code. The IR(Intermediate Representation) is closer to the source code rather than the machine code. Thus it is much more stable than the monkey solution.

In conclusion, `xgo` and monkey are compared as the following:
||xgo|monkey|
|-|-|-|
|Technique|IR|ASM|
|Function Mock|Y|Y|
|Unexported Function Mock|Y|N|
|Per-Instance Method Mock|Y|N|
|Per-Goroutine Mock|Y|N|
|Per-Generic Type Mock|Y|Y|
|Var Mock|Y|N|
|Const Mock|Y|N|
|Closure Mock|Y|Y|
|Stack Trace|Y|N|
|General Trap|Y|N|
|Compatiblility|NO LIMIT|limited to amd64 and arm64|
|API|simple|complex|
|Integration Effort|easy|hard|

# Contribution
Want to help contribute to `xgo`? Great! Check [CONTRIBUTING
](CONTRIBUTING.md) for help.

# Evolution of `xgo`
`xgo` is the successor of the original [go-mock](https://github.com/xhd2015/go-mock), which works by rewriting go code before compile.

The strategy employed by `go-mock` works well but causes much longer build time for larger projects due to source code explosion.

However, `go-mock` is remarkable for it's discovery of Trap, Trace besides Mock, and additional abilities like trapping variable and disabling map randomness.

It is the shoulder which `xgo` stands on.
