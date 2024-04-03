# xgo
[![Go Reference](https://pkg.go.dev/badge/github.com/xhd2015/xgo.svg)](https://pkg.go.dev/github.com/xhd2015/xgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/xhd2015/xgo)](https://goreportcard.com/report/github.com/xhd2015/xgo)
[![Go Coverage](https://img.shields.io/badge/Coverage-82.6%25-brightgreen)](https://github.com/xhd2015/xgo/actions)
[![CI](https://github.com/xhd2015/xgo/workflows/Go/badge.svg)](https://github.com/xhd2015/xgo/actions)
[![Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

**[English](./README.md) | 简体中文**

允许对`go`的函数进行拦截, 并提供Mock和Trace等工具帮助开发者编写测试和快速调试。

`xgo`作为一个预处理器工作在`go run`,`go build`,和`go test`之上。

`xgo`对源代码和IR(中间码)进行预处理之后, 再调用`go`进行后续的编译工作。通过这种方式, `xgo`实现了一些在`go`中缺乏的能力。

这些能力包括:
- [Trap](#trap)
- [Mock](#mock)
- [Trace](#trace).

更多细节, 参见[快速开始](#快速开始)和[文档](./doc)。

# 安装
```sh
# macOS和Linux(包括 WSL)
curl -fsSL https://github.com/xhd2015/xgo/raw/master/install.sh | bash

# Windows
powershell -c "irm github.com/xhd2015/xgo/raw/master/install.ps1|iex"
```

如果你已经安装了`xgo`, 可以通过下面的命令进行升级:

```sh
xgo upgrade
```

如果你想使用最新的版本或者本地构建, 可以使用下面的命令:

```sh
git clone https://github.com/xhd2015/xgo
cd xgo
go run ./script/build-release --local
```

验证:
```sh
xgo version
# 输出:
#   1.0.x
```

# 使用须知
`xgo` 需要`go1.17`及以上的版本。

对OS和Arch没有限制, `xgo`支持所有`go`支持的OS和Arch。

OS:
- macOS
- Linux
- Windows (+WSL)
- ...

Arch:
- x86
- x86_64(amd64)
- arm64
- ...

# 快速开始
我们基于`xgo`编写一个单元测试:

1. 确保`xgo`已经安装了, 参考[安装](#安装), 通过下面的命令验证:
```sh
xgo version
# 输出
#   1.0.x
```
如果未找到`xgo`, 你可能需要将`~/.xgo/bin`添加到你的环境变量中。

2. 创建一个demo工程:
```sh
mkdir demo
cd demo
go mod init demo
```
3. 将下面的内容添加到文件`demo_test.go`中:
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
4. 获取`xgo/runtime`的依赖:
```sh
go get github.com/xhd2015/xgo/runtime
```
5. 测试代码:
```sh
# 注意: 首次运行xgo可能会花费一点时间来
# 复制文件。后续的运行速度和go一样快
xgo test -v ./
```

输出:
```sh
=== RUN   TestFuncMock
--- PASS: TestFuncMock (0.00s)
PASS
ok      demo
```

如果使用go运行上面的代码, 它会失败:
```sh
go test -v ./
```

输出:
```sh
WARNING: failed to link __xgo_link_on_init_finished(requires xgo).
WARNING: failed to link __xgo_link_on_goexit(requires xgo).
=== RUN   TestFuncMock
WARNING: failed to link __xgo_link_set_trap(requires xgo).
WARNING: failed to link __xgo_link_init_finished(requires xgo).
    demo_test.go:21: expect MyFunc() to be 'mock func', actual: my func
--- FAIL: TestFuncMock (0.00s)
FAIL
FAIL    demo    0.771s
FAIL
```

上面的示例代码可在[doc/demo](./doc/demo)中找到.

# API
## Trap
Trap允许对几乎所有函数进行拦截, 它是`xgo`的核心机制, 是其他功能, 如Mock和Trace的基础。

下面的例子对函数的执行进行拦截并打印日志:

(查看 [test/testdata/trap/trap.go](test/testdata/trap/trap.go) 获取更多细节.)
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

使用`go`运行:

```sh
go run ./
# 输出:
#   A
#   B
```

使用`xgo`运行:

```sh
xgo run ./
# 输出:
#   trap A
#   A
#   abort B
```

`AddInterceptor()`将一个拦截器添加到全局或当前Goroutine,取决于当前是在`init`中还是`init`完成后:
- `init`中: 对所有Goroutine中的函数拦截都生效,
- `init`完成后: 仅对当前Goroutine生效, 并且在当前Goroutine退出后清理.

当在`init`完成之后调用`AddInterceptor()`, 它还会返回一个额外的清理函数, 用于提前清理设置的拦截器.

例子:

```go
func main(){
    clear := trap.AddInterceptor(...)
    defer clear()
    ...
}
```

Trap还提供了一个`Direct(fn)`的函数, 用于跳过拦截器, 直接调用到原始的函数。

## Mock
Mock简化了设置拦截器的步骤, 并允许仅对特定的函数进行拦截。

> API更多细节: [runtime/mock/README.md](runtime/mock)

Mock的API:
- `Mock(fn, interceptor)`

快速参考:
```go
// 包级别函数
mock.Mock(SomeFunc, interceptor)

// 实例方法
// 只有实例`v`的方法会被mock
// `v`可以是结构体或接口
mock.Mock(v.Method, interceptor)

// 参数类型级别的范型函数
// 只有`int`参数的才会被mock
mock.Mock(GenericFunc[int], interceptor)

// 实例和参数级别的范型方法
v := GenericStruct[int]
mock.Mock(v.Method, interceptor)

// 闭包
// 虽然很少需要mock,但支持
mock.Mock(closure, interceptor)
```

参数:
- 如果`fn`是普通函数(即包级别的函数, 类型的函数或者匿名函数(是的, xgo支持Mock匿名函数)), 则所有对该函数的调用都将被Mock,
- 如果`fn`是方法(即实例或接口的方法, 例如`file.Read`), 则只有绑定的实例会被拦截, 其他实例不会被Mock.

影响范围:
- 如果`Mock`在`init`过程中调用, 则改函数在所有的Gorotuine的调用都会被Mock,
- 否则, `Mock`在`init`完成之后调用, 则仅对当前Goroutine生效, 其他Goroutine不受影响.

拦截器签名: `func(ctx context.Context, fn *core.FuncInfo, args core.Object, results core.Object) error`
- 如果返回`nil`, 则目标函数被Mock,
- 如果返回`mock.ErrCallOld`, 则目标函数会被调用,
- 否则, 返回的其他错误, 会被设置为目标函数的error返回值, 目标函数不会被调用。

Mock还有两个额外的API, 它们基于名称进行拦截, 更多细节，参见[runtime/mock/README.md](runtime/mock/README.md)。

方法Mock示例:
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

    // myStruct受影响
    name := myStruct.Name()
    if name!="mock struct"{
        t.Fatalf("expect myStruct.Name() to be 'mock struct', actual: %s", name)
    }

    // otherStruct不受影响
    otherName := otherStruct.Name()
    if otherName!="other struct"{
        t.Fatalf("expect otherStruct.Name() to be 'other struct', actual: %s", otherName)
    }
}
```

**关于标准库Mock的注意事项**: 出于性能和安全考虑, 标准库中只有一部分包和函数能被Mock, 这个List可以在[runtime/mock/stdlib.md](./runtime/mock/stdlib.md)找到. 如果你需要Mock的标准库函数不在列表中, 可以在[Issue#6](https://github.com/xhd2015/xgo/issues/6)中进行评论。

## Patch
`runtime/mock`还提供了另一个API:
- `Patch(fn,replacer) func()`

参数:
- `fn` 与[Mock](#mock)中的第一个参数相同
- `replacer`一个用来替换`fn`的函数

注意: `replacer`应当和`fn`具有同样的签名。

返回值:
- 一个`func()`, 用来提前移除`replacer`

Patch将`fn`替换为`replacer`,这个替换仅对当前goroutine生效.在当前Goroutine退出后, `replacer`被自动移除。

例子:
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

## Trace
在调试一个非常深的调用栈时, 通常会感觉非常痛苦, 并且效率低下。

为了解决这个问题, Trace实现了结构化的调用堆栈信息收集。在Trace的帮助下, 通常不需要很多Debug即可知道函数调用的细节:

(代码参见[test/testdata/trace/trace.go](test/testdata/trace/trace.go))
```go
package main

import (
    "fmt"

    "github.com/xhd2015/xgo/runtime/trace"
)

func init() {
    trace.Enable()
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

使用`go`运行:

```sh
go run ./
# 输出:
#   A
#   B
#   C
#   C
```

使用`xgo`运行:

```sh
XGO_TRACE_OUTPUT=stdout xgo run ./
# 输出一个JSON:
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
# NOTE: 省略其他的字段, 只展示关键的信息
```

使用`xgo`自带的可视化工具查看Trace: `xgo tool trace TestExample.json`

输出:
![trace html](cmd/trace/testdata/stack_trace.jpg "Trace")

默认情况下, Trace会在当前目录下写入堆栈记录, 可通过环境变量`XGO_TRACE_OUTPUT`进行控制:
- `XGO_TRACE_OUTPUT=stdout`: 堆栈记录会被输出到stdout, 方便debug,
- `XGO_TRACE_OUTPUT=<dir>`: 堆栈记录被写入到`<dir>`目录下,
- `XGO_TRACE_OUTPUT=off`: 关闭堆栈记录收集。

# 实现原理

> 仍在整理中...

参见[Issue#7](https://github.com/xhd2015/xgo/issues/7)

# 为何使用`xgo`?
原因很简单: **避免**interface.

是的, 如果仅仅是为了Mock而抽象出一个interface, 然后在别的地方注入不同的实例, 这让我感觉很疲惫，这也不是事情正确方式。

基于interface的Mock强制我们按照一种唯一的方式写代码: 就是基于接口。基于自由的原因, 我反对这一点。

Monkey patch是Mock最简单直接的答案。但是现有的Monkey库存在很多兼容性问题。

所以, 我创建了`xgo`, 并希望`xgo`能够统一解决go中Mock的所有问题。

# 对比`xgo`和`monkey`
Bouk创建了[bouk/monkey](https://github.com/bouk/monkey)工程, 基于他的博客 https://bou.ke/blog/monkey-patching-in-go.

简单来说, 该工程使用了低级汇编代码在运行时替换函数地址, 从而实现对函数的拦截。但是, 这带来了非常多的问题, 尤其是在MacOS上。

随着新的go版本和Apple新的芯片发布, 该工程就不再被维护, 停止迭代。然后两个工程继承了monkey的思想, 继续基于ASM, 添加对新的go版本和芯片(如Apple M1)的支持。

但是一如既往, 它们并没有解决ASM带来的底层的兼容性问题, 比如不能跨平台, 需要往只读代码段写入内容, 以及缺乏通用的拦截器。

所以, go的开发者们不得不频繁面对Mock失效的问题。

Xgo通过在IR(Intermediate Representation)重写来成功地避开了这些问题, 因为IR更靠近源代码而不是机器代码。这带来了比monkey更加稳定可靠, 以及可持续的解决方案。 

总而言之, `xgo`和monkey对比如下:
||xgo|monkey|
|-|-|-|
|底层实现|IR|ASM|
|函数Mock|Y|Y|
|未导出的函数Mock|Y|N|
|实例级别方法Mock|Y|N|
|Goroutine级别Mock|Y|N|
|范型特定实例Mock|Y|Y|
|闭包/匿名函数Mock|Y|Y|
|堆栈收集|Y|N|
|通用拦截器|Y|N|
|兼容性|NO LIMIT|仅支持amd64和arm64|
|API|简单|复杂|
|集成使用和维护|简单|困难|

# 参与贡献
希望参与到`xgo`的开发吗? 非常好, 可以查看[CONTRIBUTING
](CONTRIBUTING.md)获取更多细节。

# `xgo`的演化
`xgo`基于[go-mock](https://github.com/xhd2015/go-mock)演化而来, 后者基于编译时代码重写.

尽管go-mock的代码重写能够完美工作, 但是由于重写后代码膨胀, 导致编译速度变慢, 在大型工程上尤为明显。

不过, go-mock仍然是值得注意的, 因为正是基于代码重写, 才引入了Trap, Trace和Mock这些概念, 以及其他的能力, 比如变量拦截, Map随机化关闭等。

`xgo`正是在go-mock的基础上重新设计的。
