# LC_UUID Missing Error on macOS 26+ (arm64)

## Symptom

When running xgo (compiled by Go < 1.23) on macOS 26+ (Apple Silicon), dyld crashes:

```
dyld[84494]: missing LC_UUID load command in /private/var/folders/.../go-build.../b001/exe/xgo
dyld[84494]: missing LC_UUID load command
signal: abort trap
FAIL go-release/go1.22.12: exit status 1
```

## Root Cause

On macOS 26+, dyld requires all Mach-O executables to have an `LC_UUID` load command. The Go internal linker added `LC_UUID` emission in Go 1.22.9 / 1.23.3 / 1.24+. However, empirically go1.22.12 still produces binaries without `LC_UUID` on some systems, causing the crash.

## Error Chain

When running tests via `script/run-test --include go1.22.12`:

1. `script/run-test` runs `go1.22.12/bin/go run ./cmd/xgo test ...`
2. `go1.22.12/bin/go` compiles the xgo binary to `/var/folders/.../go-build.../exe/xgo`
3. The compiled xgo binary lacks `LC_UUID` (go1.22.12's linker doesn't emit it)
4. dyld refuses to execute the xgo binary → `signal: abort trap`

The same issue affects any binary built by Go < 1.23 on macOS 26+, including:
- The xgo binary itself (built by `go run ./cmd/xgo`)
- Toolchain binaries rebuilt by xgo (go, compile, cover)
- End-user test binaries

## Fix: External Linker (`-ldflags=-linkmode=external`)

Using the system linker (clang/ld) via `-linkmode=external` ensures `LC_UUID` is always present, regardless of the Go version.

### Changes

**`instrument/build/build.go`** — `NeedExternalLinker()` (line 63):
- Changed version check from `goVersion.Minor < 22` to `goVersion.Minor < 23`
- Covers all Go 1.22.x which empirically still fails despite Go 1.22.9+ having the fix in theory

**`instrument/build/build.go`** — already protected toolchain binaries (go, compile, cover) rebuilt by xgo via `RebuildGoBinary`, `RebuildGoToolCompile`, `RebuildGoToolCover`.

**`cmd/xgo/main.go`** — already protected end-user build/test/run commands (line 917).

**`script/run-test/main.go`** — added protection for the xgo binary itself:
- `doRunTest()`: adds `-ldflags=-linkmode=external` to `go run ./cmd/xgo` and `go build ./cmd/xgo` commands based on goroot's Go version (via `Opts.GoVersion`)
- `resetInstrument` block: adds `-ldflags=-linkmode=external` on darwin/arm64
- `setupGoroot()`: adds `-ldflags=-linkmode=external` on darwin/arm64

### Why `setupGoroot` and `resetInstrument` use runtime checks

These paths use the system `go` from PATH, not the goroot's `go`. The goroot's Go version is not available at call time, so they defensively add external linker flags whenever host is `darwin/arm64`.

## References

- Go issue: https://github.com/golang/go/issues/68678, #78012
- Initial fix commit: `b0c14c5` — *"instrument: add LC_UUID support for darwin/arm64"*
- Refactor commit: `b5223c8` — *"refactor: add ExternalLinkerFlags and host helpers"*
- This gap fix: extends `NeedExternalLinker` to `< 1.23` and wires flags into `script/run-test`
