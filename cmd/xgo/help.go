package main

const help = `
Xgo is a tool for instrumenting Go source code.

Xgo works as a drop-in replacement for these go commands: 'go build','go run' and 'go test'. 
So flags accepted by these commands are also accepted by xgo.

Usage:
    xgo <command> [arguments]

The commands are:
    build       build instrumented code, extra arguments are passed to 'go build' verbatim
    run         run instrumented code, extra arguments are passed to 'go run' verbatim
    test        test instrumented code, extra arguments are passed to 'go test' verbatim
    exec        execute a command verbatim
    version     print xgo version
    revision    print xgo revision
    upgrade     upgrade to latest version of xgo
    tool        invoke xgo tools   

Examples:
    xgo build -o main ./                         build current module
    xgo build -o main -gcflags="all=-N -l" ./    build current module with debug flags
    xgo run ./                                   run current module
    xgo test ./...                               test all test cases of current module
    xgo exec go version                          print instrumented go version
    xgo tool help                                print help for xgo tools

Examples of Trace:
    xgo test -run TestSomething --strace ./      test and collect stack trace
    xgo tool trace TestSomething.json            view collected stack trace

Example of Test Explorer:
    xgo e                                        open test explorer, alias for xgo tool test-explorer
    xgo explorer                                 alias for xgo tool test-explorer
    xgo test-explorer                            alias for xgo tool test-explorer
    xgo tool test-explorer                       canonical command to open test-explorer

See https://github.com/xhd2015/xgo for documentation.

`
