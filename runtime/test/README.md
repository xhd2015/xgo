# Tests of the runtime package
Runnable tests are listed in file [../../script/run-test/main.go](../../script/run-test/main.go).

# Run tests
```sh
# all

# patch
go run -tags dev ./cmd/xgo test --with-goroot go1.19.13 --project-dir runtime/test ./patch

# trace
go run -tags dev ./cmd/xgo test --with-goroot go1.19.13 --project-dir runtime/test ./trace

# specific test
go run -tags dev ./cmd/xgo test --with-goroot go1.17.13 --project-dir ./runtime/test/mock_var -v -run TestThirdPartyTypeMethodVar

```

# Debug tests
```bash
# build
go run -tags dev ./cmd/xgo test --with-goroot go1.19.13 -c -o __debug_bin_test -gcflags="all=-N -l" --project-dir runtime/test ./patch

# debug
dlv exec --listen=:2345 --api-version=2 --check-go-version=false --headless -- ./__debug_bin_test -test.run TestPatchTypeMethodCtxArg
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