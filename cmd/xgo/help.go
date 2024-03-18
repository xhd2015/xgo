package main

// TODO: revise documentation link
const help = `
Xgo is a tool for instrumenting Go source code.

Xgo works as a drop-in replacement for these go commands: 'go build','go run' and 'go test'. 
So flags accepted by these commands are also accepted by xgo.

Usage:
    xgo <command> [arguments]

The commands are:
    build       build instrumented code, extra arguments are passed 'go build' verbatim
    run         run instrumented code, extra arguments are passed 'go run' verbatim
    test        test instrumented code, extra arguments are passed 'go test' verbatim
    exec        execute a command verbatim
    version     print xgo version
    revision    print xgo revision

Examples:
    xgo buil -o main ./                          build current module
    xgo buil -o main -gcflags="all=-N -l" ./     build current module with debug flags
    xgo run ./                                   run current module
    xgo exec go version                          print instrumented go version

See https://github.com/xhd2015/xgo for documentation.

`
