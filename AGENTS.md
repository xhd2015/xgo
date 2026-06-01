# AGENTS.md

## Project Overview

xgo is a Go tool that instruments Go code for testing, mocking, tracing, and intercepting. It acts as a drop-in replacement for `go` commands (build, test, run, etc.).

## Clean `go.mod`

Both `go.mod` and `runtime/go.mod` are very thin, has no any dependencies.

That is delibrate, because `xgo` serve as a foundamental driver for go testing, so it should not introduce any dependency burden to end user.

Please always keep them clean.

## Testing

Tests are located under `runtime/test/` and are organized as standalone Go modules (each with its own `go.mod`, replacing `github.com/xhd2015/xgo/runtime` with the local `../../..`). They require the xgo instrumentation to work, so they must be run via the xgo test driver.

### Test Driver

Tests are driven by `script/run-test/`:

```bash
# Run all tests:
go run ./script/run-test

# List all discovered tests:
go run ./script/run-test list

# Run a specific test module:
go run ./script/run-test ./runtime/test/trap/interceptor/interceptor_ctx

# Run with specific test pattern and verbose output:
go run ./script/run-test -run TestHelloWorld -v ./runtime/test/trap/interceptor/interceptor_ctx

# Run against a specific GOROOT (from go-release/):
go run ./script/run-test --include go1.21.8 ./runtime/test/trap/interceptor/interceptor_ctx
```

### How It Works

1. `scanGoMods` recursively discovers directories containing `go.mod` under `runtime/test/`
2. Each discovered module becomes a `TestConfig` with `Dir` set to the module path
3. The driver runs `go run ./cmd/xgo test <module>` for each test config (or `go test` if `UsePlainGo: true`)
4. Modules can have an optional `test-config.txt` file to customize test behavior

### test-config.txt

Put a `test-config.txt` in the test module directory to configure:
```
flags: -v -count=1          # extra flags passed to go test
args: ./...                 # which packages to test
use-plain-go: true          # use go test instead of xgo test
vendor-if-missing: true     # run go mod vendor before testing
env: GOOS= GOARCH=          # environment variables
skip: <reason>              # skip this test
```

### Running a Single Test

```bash
# method 1: specific test module + test pattern
go run ./script/run-test -run TestFuncCustomCtx -v ./runtime/test/trap/interceptor/interceptor_ctx

# method 2: use --include for a specific GOROOT from go-release/
go run ./script/run-test --include go1.21.8 -run TestFuncCustomCtx -v ./runtime/test/trap/interceptor/interceptor_ctx

# method 3: run all tests under runtime/test/
go run ./script/run-test ./runtime/test/all

# Plain go tests (no xgo instrumentation, for packages without trap/mock deps):
go run ./script/run-test ./...
```

### Test Structure Convention

Each test module under `runtime/test/`:
- Has its own `go.mod`
- Has `replace github.com/xhd2015/xgo/runtime => ../../..` to use the local runtime
- Test files (in same package) provide the functions being tested
- Interceptors/mocks are set up in test functions via `trap.AddInterceptor` etc.

## Debugging

When encountering hard bugs, read `GOCHAS.md` for inspiration.

## Code Organization

| Directory | Purpose |
|-----------|---------|
| `cmd/xgo/` | The xgo CLI tool |
| `runtime/` | The xgo runtime library (trap, mock, trace, core) |
| `runtime/internal/trap/` | Low-level trap implementation |
| `runtime/test/` | Integration tests (standalone modules, xgo-instrumented) |
| `instrument/patch/` | Patch DSL parser/engine for instrumenting Go stdlib (see [PATCH_DSL.md](PATCH_DSL.md)) |
| `patches/` | Per-version `.xgo.patch` files for GOROOT instrumentation |
| `script/` | Build and test scripts |
| `script/run-test/` | Test driver |
| `script/download-go/` | Downloads Go releases for cross-version testing |
| `test/` | Additional test infrastructure |
