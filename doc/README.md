# Overview of `xgo`
`xgo` is a wrapper around `go`.

All magic is behind the go's `-toolexec` flag.

Run `go help build`, and you will find this flag:
```
-toolexec 'cmd args'
        a program to use to invoke toolchain programs like vet and asm.
        For example, instead of running asm, the go command will run
        'cmd args /path/to/asm <arguments for asm>'.
        The TOOLEXEC_IMPORTPATH environment variable will be set,
        matching 'go list -f {{.ImportPath}}' for the package being built.
```

When you run `xgo build ./my/example`, it does the following things:
 1. find the GOROOT,
 2. copy the GOROOT into ~/.xgo/go-instruments/GOROOT to prepare for instrumenting,
 3. apply patch to ~/.xgo/go-instruments/GOROOT, both for compiler and runtime,
 4. build the instrumented compiler,
 5. invoke go build with extra flag: `go build -toolexec exec_tool ./my/example`,
 6. the `exec_tool` then forward all compile command to the instrumented compiler 
 7. once all compilation finished, go invoke link to generate the executable, and you get a instrumented binary!