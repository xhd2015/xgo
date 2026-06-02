# Why Double Setup (--reset-instrument) Ever Failed

## The Bug

Running `xgo setup --reset-instrument` on an already-instrumented GOROOT would fail with:

- **Linux AMD64 (CI):** `undefined: src.NoXPos_xgo_get` in `cmd/internal/obj`
- **macOS ARM64:** `go: go.mod requires go >= 1.25 (running go 1.24.11; GOTOOLCHAIN=local)`

Both errors originated from the same root cause: the `rebuild-compiler` generate step did not set `GO_BYPASS_XGO=true`.

## The Chain of Events

### 1. The instrumented go binary has an xgoPrecheck hook

During GOROOT instrumentation, xgo patches `src/cmd/go/main.go` (via `xgo_main.go`) to add an `xgoPrecheck` call that intercepts `go build`, `go test`, and `go run`:

```go
// instrument/instrument_go/xgo_main_template.go:19
if os.Getenv("GO_BYPASS_XGO") == "true" {
    return false  // no interception
}
// delegate to xgo
xgoCmd := exec.Command("xgo", args...)
xgoCmd.Env = append(os.Environ(), "GO_BYPASS_XGO=true")
```

This hook is compiled into the instrumented `bin/go` binary. When `GO_BYPASS_XGO` is not set and `xgo` is found in `PATH`, the instrumented go delegates **all** build/test/run commands back to xgo.

### 2. The rebuild-compiler step invokes the instrumented go

The `__config__.json` for file-based patches defines a `rebuild-compiler` step:

```json
{
  "kind": "rebuild-compiler",
  "cmd": "${INSTRUMENT_GOROOT}/bin/go build -a -o ${INSTRUMENT_GOROOT}/pkg/tool/${GOOS}_${GOARCH}/compile cmd/compile"
}
```

`${INSTRUMENT_GOROOT}/bin/go` is the **already-instrumented** go binary. Without `GO_BYPASS_XGO=true`, the `xgoPrecheck` hook fires.

### 3. xgo delegates recursively, triggering instrumentation on toolchain packages

The instrumented go spawns `xgo build --go build -a ... cmd/compile`. xgo processes this as a normal user build command:

1. Loads all transitive dependencies of `cmd/compile`
2. Calls `CheckInstrument` for each package
3. `CheckInstrument("cmd/internal/obj")` returns `allow=true` — `PkgWithinModule("cmd/internal/obj", "internal")` is `false` because the path starts with `"cmd/"`, not `"internal/"`
4. Variable trapping rewrites `src.NoXPos` → `src.NoXPos_xgo_get()` — but the getter function is not properly generated for stdlib packages
5. The build fails with `undefined: src.NoXPos_xgo_get`

On macOS, the error manifests differently (GOTOOLCHAIN version mismatch) because the xgo binary runs under the bootstrapping Go version, which may differ from the target GOROOT version.

### 4. Why `--reset-instrument` specifically triggers it

The `--reset-instrument` flag on an already-instrumented GOROOT leaves the existing instrumented files in place (because `instrumented=true` is detected via `xgo-revision.txt`). The `rebuild-compiler` step then runs the instrumented `bin/go` — the one with the `xgoPrecheck` hook.

A fresh first-time setup does NOT hit this because `bin/go` is the original (uninstrumented) binary at that point.

**Reproduction requires passing the INSTRUMENTED goroot path to `--reset-instrument`**, not the original goroot. The test at `test/integrations/xgo-with-setup-repro/main.go` demonstrates this.

## The Fix

Add `GO_BYPASS_XGO=true` to the environment of generate commands that invoke the instrumented go binary.

### In `__config__.json`

```json
{
  "kind": "rebuild-compiler",
  "comments": [
    "rebuild compiler with patched source",
    "GO_BYPASS_XGO prevents the instrumented go binary's xgoPrecheck hook from delegating back to xgo"
  ],
  "env": {"GO_BYPASS_XGO": "true"},
  ...
}
```

Applied to all three rebuild entries: `rebuild-compiler`, `rebuild-stdlib`, `rebuild-go`.

### In `apply.go`

The `GenerateEntry` struct gained an `Env` field (`map[string]string`). During generate command execution (`apply.go:118`), entries from `gen.Env` are appended to the subprocess environment after the hardcoded vars (`GOROOT`, `GOTOOLCHAIN=local`, `GOOS`, `GOARCH`).

### Historical note

The legacy programmatic instrumentation path (`instrument/build/build.go:29`) already set `GO_BYPASS_XGO=true`. The file-based patch path (`instrument/patch/apply.go:118`) simply missed it when it was first implemented.

## Code References

| File | Line | Role |
|------|------|------|
| `instrument/patch/apply.go` | 31-38 | `GenerateEntry` struct with `Env` field |
| `instrument/patch/apply.go` | 118-123 | Env var assembly for generate commands |
| `patches/go1.25/__config__.json` | 17-21 | `rebuild-compiler` entry with `env` |
| `instrument/instrument_go/xgo_main_template.go` | 14-31 | `xgoPrecheck` — checks `GO_BYPASS_XGO` |
| `instrument/build/build.go` | 29 | Legacy path already had `GO_BYPASS_XGO=true` |
| `test/integrations/xgo-with-setup-repro/main.go` | — | Minimal reproduction test |
