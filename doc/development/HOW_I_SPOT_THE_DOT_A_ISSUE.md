# How I Spot the `.a` Issue

When `functab.GetFuncs()` returns 0 entries but the compiler is correctly built and link rewriting works, the problem is almost always stale pre-compiled `.a` files.

## The Mechanism

Go ships pre-compiled standard library `.a` files in `pkg/$GOOS_$GOARCH/`. When you run `go build` or `go test` without `-a`, Go links against these cached `.a` files **without checking whether the source has changed**. So patching `src/runtime/xgo_trap.go` (adding `XgoRegister`) has no effect — the `runtime.a` from the original GOROOT is still used.

At runtime, the test binary's `functab.init()` calls `runtime.XgoSetupRegisterHandler()`, which exists in the patched source but is **absent from the stale `runtime.a`**. The call becomes a no-op, and subsequently `XgoRegister()` calls from init functions land nowhere. Result: 0 functab entries.

## Spot It in 3 Steps

### 1. Check if the symptom matches

- ✅ Compiler binary has patch code: `strings compile | grep XgoRegister` returns matches
- ✅ Debug prints show `LinkXgoInit` finds `__xgo_init_*` and rewrites `__xgo_register_* → XgoRegister`
- ❌ But `functab.GetFuncs()` returns 0 entries

This pattern means the pipeline is correct at compile time but breaks at link/runtime time. The linker resolved `XgoRegister` against a stale `runtime.a`.

### 2. Check the test binary for XgoRegister

```sh
go test -c -o /tmp/test.bin ./...
strings /tmp/test.bin | grep XgoRegister
```

If this returns nothing, the test binary's runtime does not have `XgoRegister`.

Contrast with the compiler binary:
```sh
strings "$INST/pkg/tool/darwin_arm64/compile" | grep XgoRegister
```

This should return matches (the compiler was rebuilt from patched source).

### 3. Compare runtime.a timestamps

```sh
INST=$(ls -d ~/.xgo/go-instrument/go1.25.10_*/go1.25.10)
ls -la "$INST/pkg/darwin_arm64/runtime.a"
ls -la go-release/go1.25.10/pkg/darwin_arm64/runtime.a
```

If they have the same timestamp and size, the instrumented GOROOT's `runtime.a` was **never rebuilt** — it's still the original.

## The Fix

Add a `rebuild-stdlib` step to `patches/go<VERSION>/__config__.json` that runs after `rebuild-compiler`:

```json
{
    "kind": "rebuild-stdlib",
    "cwd": "${INSTRUMENT_GOROOT}/src",
    "cmd": "${INSTRUMENT_GOROOT}/bin/go install std",
    "comment": "rebuild all standard library .a files from patched source"
}
```

`go install std` compiles every standard library package from source and writes fresh `.a` files into `pkg/$GOOS_$GOARCH/`, replacing the stale originals.

## Why the Programmatic Path Doesn't Have This Problem

The old `V1_0_0` path passes `-a` to `go test` (because `compilerChanged` is true after rebuilding the compiler). The `-a` flag forces Go to recompile everything from source at test time, so the stale `.a` files are never used. The file-based patch path does not set `-a`, so `go install std` at setup time is the correct fix.
