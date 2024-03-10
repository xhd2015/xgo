# Play with go compiler
This directory demonstrates how to play with the go compiler.
It modifies the compiler so that it inserts two extra print:
 - by modifying AST: hello Syntax
 - by modifying IR: hello IR
 - by modifying IR and inserting __x_trap to runtime: hello Trap

```sh
$ ./debug.sh build-comipler
$ ./debug.sh build
$ ./main.bin 
hello IR
hello Trap
hello Syntax
hello world
```

Current go version is go1.22.

# Debug the compiler
Compiler entrance: [../compile/main.go](../compile/main.go)

```sh
./debug.sh build-compiler
./debug.sh debug # this will hang the terminal, and you can copy the output configuration to .vscode
```

The compiler should be `compile-devel` or `compile`? That depends on go version.
Because `compile-devel -V=full` outputs `compile-devel version go1.21.7`,

- With go1.22, that is accepted.

- With go1.20, the `compile-devel` part does not match `compile`,so failed.
  The message is:
```
go: parsing buildID from go tool compile -V=full: unexpected output:
        compile-devel version go1.20.14
```
Emitted from `src/cmd/go/internal/work/buildid.go`.

# Implement a mock framework
With this techique it is trivil to implement a mock framework.

Just use the `trap.AddInterceptor({Pre:..., Post:...})` to define pre to detect function point in interested, and return a `trap.ErrAbort` to skip the function.

See [./test/test_trap/mock/mock.go](./test/test_trap/mock/mock.go):
```go
trap.AddInterceptor(trap.Interceptor{
    Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (interface{}, error) {
        if strings.Contains(f.FullName, "testArgs") {
            fmt.Printf("Mock: %s\n", f.FullName)
            p := args.Results[0].(*int)
            *p = 20
            return nil, trap.ErrAbort
        }
        return nil, nil
    }
})
```

# Implement a stack trace collector
It is possible to collect a runtime stack trace collector, see [core/trace/trace.go](core/trace/trace.go).
Usage:
```sh
./with-go-devel ./debug.sh build -v -gcflags \"all=-N -l\" -o ./test_trap.bin ./test/test_trap
```
Or just use the vscode launch config `Launch test_trap`.


Example:
```json
{
    "Children": [
        {
            "FuncInfo": {
                "FullName": "main.main"
            },
            "Children": [
                {
                    "FuncInfo": {
                        "FullName": "main.testArgs"
                    },
                    "Recv": null,
                    "Args": [
                        "a"
                    ],
                    "Results": [
                        1
                    ],
                    "Children": [
                        {
                            "FuncInfo": {
                                "FullName": "main.num.add"
                            },
                            "Recv": 1,
                            "Args": [
                                2
                            ],
                            "Results": null,
                            "Children": [
                                {
                                    "FuncInfo": {
                                        "FullName": "github.com/xhd2015/xgo/runtime/pkg.Hello"
                                    },
                                    "Recv": null,
                                    "Args": [
                                        "pkg"
                                    ],
                                    "Results": null,
                                    "Children": [
                                        {
                                            "FuncInfo": {
                                                "FullName": "github.com/xhd2015/xgo/runtime/pkg.Mass.Print"
                                            },
                                            "Recv": 1,
                                            "Args": [
                                                "g"
                                            ],
                                            "Results": null,
                                            "Children": [
                                                {
                                                    "FuncInfo": {
                                                        "FullName": "github.com/xhd2015/xgo/runtime/pkg.(*Person).Greet"
                                                    },
                                                    "Recv": {
                                                        "Name": "test"
                                                    },
                                                    "Args": [
                                                        "runtime"
                                                    ],
                                                    "Results": null,
                                                    "Children": [
                                                        {
                                                            "FuncInfo": {
                                                                "FullName": "github.com/xhd2015/xgo/runtime/pkg.Hello.func1"
                                                            }
                                                        }
                                                    ]
                                                }
                                            ]
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

# Development
## How to add customized function to runtime?
1.Edit [../compile/internal/typecheck/_builtin/runtime.go](../compile/internal/typecheck/_builtin/runtime.go) to add function declaration,
2.Execute go generate
```sh
./debug.sh gen-runtime-type
# or
./with-go-devel.sh go generate ../compile/internal/typecheck
```

Supported flags:
- `-stdout`  output to stdout instead of file
- `-nofmt` skip formatting

See [../compile/internal/typecheck/mkbuiltin.go](../compile/internal/typecheck/mkbuiltin.go)

Also:
```sh
(cd ../compile/internal/typecheck; ../../../test/with-go-devel.sh go run ./mkbuiltin.go -stdout -nofmt)
```

## Check runtime symbols
```sh
./with-go-devel.sh go tool nm runtime.a
```

# Configure git exclude
```
root=$(git rev-parse --show-toplevel)
mkdir -p "$root/.git/info"
cat >>"$root/.git/info/exclude" <<'EOF'
/src/cmd/test/*.log
/src/cmd/test/compile-devel
/src/cmd/test/*.bin
/src/cmd/test/*.a
EOF
```