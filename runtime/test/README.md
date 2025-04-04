# Tests of the runtime package
Runnable tests are listed in file [../../script/run-test/main.go](../../script/run-test/main.go).

All tests must be run by xgo.

`runtime/test` has special access to `runtime/internal` functions, like `trap.Inspect()`.

# Run tests
```sh
# all

# patch
go run -tags dev ./cmd/xgo test --with-goroot go1.19.13 --project-dir ./runtime/test/patch

# trace
go run -tags dev ./cmd/xgo test --with-goroot go1.19.13 --project-dir runtime/test ./trace

# specific test
go run -tags dev ./cmd/xgo test --with-goroot go1.17.13 --project-dir ./runtime/test/mock_var -v -run TestThirdPartyTypeMethodVar

# with -cover
go run -tags dev ./cmd/xgo test --with-goroot go1.19.13 -cover --project-dir runtime/test ./patch
```

# Debug tests
```bash
# build
go run -tags dev ./cmd/xgo test --with-goroot go1.19.13 -c -o __debug_bin_test -gcflags="all=-N -l" --project-dir runtime/test ./patch

# debug
dlv exec --listen=:2345 --api-version=2 --check-go-version=false --headless -- ./__debug_bin_test -test.run TestPatchTypeMethodCtxArg

# build and run
go run -tags dev ./cmd/xgo test --with-goroot go1.19.13 --log-debug --project-dir runtime/test -c -gcflags="all=-N -l" -o __debug_bin_test ./trace/go_trace && ./__debug_bin_test -test.v -test.run TestGoTraceSync
```


VSCode launch.json:
```json
{
    "configurations": [
        {
            "name": "Debug dlv localhost:2345",
            "type": "go",
            "debugAdapter": "dlv-dap",
            "request": "attach",
            "mode": "remote",
            "port": 2345,
            "host": "127.0.0.1",
            "cwd": "./"
        }
    ]
}
```

# Debug xgo
```bash
go build -o xgo -gcflags="all=-N -l" -tags dev ./cmd/xgo

dlv exec --listen=:2345 --api-version=2 --check-go-version=false --headless --  xgo test --with-goroot go1.19.13 --project-dir runtime/test -c -gcflags="all
=-N -l" -o __debug_bin_test ./patch
```