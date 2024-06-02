package main

const help = `
go-tool-debug-compile is a tool used to debug the go compiler.

Usage:
   go-tool-debug-compile [options]

Options:
     --compile-options-file OP_FILE  read compiler options from OP_FILE, default: debug-compile.json
     --project-dir DIR               target project dir, usually leave empty is fine
     --env K=V                       env passed to compiler, can be repeated
     --compiler BINARY               instead of debugging the compiler specified from OP_FILE, use a custom go compiler binary, NOTE: the compiler should be built with -gcflags="all=-N -l"
     --run-only                      don't start a debugger, just run the compiler
  -h,--help                          show help

Options passed to compiler:
  -N
  -l
  -c C
  -cpuprofile   PROFILE
  -blockprofile PROFILE
  -memprofile   PROFILE
  -memprofilerate RATE

Run 'go tool compile --help' for these options.

Examples:
  xgo build --debug-compile ./          build current package, and generate a debug-compile.json
  go-tool-debug-compile                 start a debugger
  go-tool-debug-compile --run-only      run the compiler and print cost

See https://github.com/xhd2015/xgo for documentation.

`
