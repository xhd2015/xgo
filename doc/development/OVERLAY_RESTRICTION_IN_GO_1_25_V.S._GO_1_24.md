# Why `runtime/test/functab` Test Passes on go1.24 but Fails on go1.25

## Symptom

```
go1.24 + --use-file-patches: PASS (6 entries, correct order)
go1.25 + --use-file-patches: FAIL — 7 entries, funcInfos[0] is runtime/trace.Trace, all shifted
```

## Root Cause: Go 1.25 Banned GOMODCACHE Overlays

This can be verified via running:
```sh
go run ./script/run-test/ --include go1.25.10 ./runtime/test/build/overlay_build_cache_error_with_go
```

### The Trigger: `needLocalRuntime` Gate

**`cmd/xgo/main.go:710`**:
```go
needLocalRuntime := goVersion.Minor >= 25
```

This boolean exists because **Go 1.25 forbids `-overlay` from replacing files under `GOMODCACHE`** (commit `3123289`):

> Go 1.25 forbids overlay replacements for files under GOMODCACHE.
> Always load the xgo runtime into a local directory (.xgo/gen/modules/)
> when running with Go 1.25+ so modifications are applied to the local
> copy instead of creating forbidden GOMODCACHE overlay entries.

On go1.24, xgo modifies xgo runtime source files (like `trace.go`) via **overlay** — the overlay shadows the GOMODCACHE-resolved files. On go1.25 this doesn't work, so xgo must:

1. Copy the runtime to a local directory outside GOMODCACHE
2. Modify files directly on disk in that local directory
3. Create a `replace` directive in a modified `go.mod` (`-modfile`) to resolve runtime from local dir
4. **Add blank imports** (`import _ "runtime/trace"`) to ensure all runtime packages are compiled

### The Causal Chain

```
Go 1.25 bans overlay for GOMODCACHE
  → needLocalRuntime = true (main.go:710)
    → importRuntimeDepGenOverlay() called (main.go:713)
      → addBlankImports() called (trace.go:142, ALWAYS)
        → Add "import _ "runtime/trace"" to test files
          → runtime/trace IS compiled + linked into binary
            → Instrumented trace.go's init() runs
              → XgoRegister(&TraceFuncInfo) called
                → Trace registered in functab (7th entry, Stdlib=false)
                  → Rebuilt stdlib changes init DAG order
                    → Trace at index 0
                      → Test expects SomeVar at index 0 → ALL shifted → FAIL
```

On go1.24, `needLocalRuntime=false`, so `importRuntimeDepGenOverlay()` never runs, `addBlankImports()` never adds the blank import, `runtime/trace` is never compiled, its `init()` never runs, `Trace` never enters functab → 6 entries → PASS.

## Detailed Technical Walkthrough

### Step 1: `LinkXgoRuntime` Always Loads All Runtime Packages

**`instrument/instrument_xgo_runtime/link.go:69-78`** — whether go1.24 or go1.25, ALL xgo runtime packages are loaded for instrumentation:
```go
packages, err := load.LoadPackages([]string{
    constants.RUNTIME_INTERNAL_RUNTIME_PKG,
    constants.RUNTIME_CORE_PKG,
    constants.RUNTIME_TRAP_FLAGS_PKG,
    constants.RUNTIME_FUNCTAB_PKG,
    constants.RUNTIME_LEGACY_CORE_INFO_PKG,
    constants.RUNTIME_MOCK_PKG,
    constants.RUNTIME_TRACE_PKG,   // always loaded
    constants.RUNTIME_TRAP_PKG,
}, opts)
```

### Step 2: `LinkXgoRuntime` Instruments `trace.go`

**Lines 218-234** — traps `Trace` function, stores `funcInfos` in `traceFile.TrapFuncs`:
```go
if foundFunctabPkg && traceFile != nil {
    funcInfos, extraFuncs := instrument_func.TrapFuncs(edit, RUNTIME_TRACE_PKG, ...)
    traceFile.TrapFuncs = funcInfos
}
```

### Step 3: `registerFuncTab` Generates Registration Code for ALL Packages

**`cmd/xgo/instrument.go:348`** — no `AllowInstrument` filter:
```go
func registerFuncTab(packages *edit.Packages) {
    for _, pkg := range packages.Packages {
        // NOTE: no AllowInstrument check here!
        // The main instrument loop (line 227) has one, but this doesn't
        for _, file := range pkg.Files {
            instrument_reg.RegisterFuncTab(fset, file, pkgPath, stdlib)
        }
    }
}
```

Unlike the main instrument loop at line 227 (`if !pkg.AllowInstrument { continue }`), `registerFuncTab` processes ALL packages, including xgo runtime packages where `AllowInstrument=false`.

### Step 4: Functab Registration Code Generated for Trace

`RegisterFuncTab` generates Go code like:
```go
var __xgo_func_info_0_0 = &XgoFuncInfo{Kind:0, Pkg:__xgo_pkg_0, Name:"Trace", IdentityName:"Trace", ...}
func init() { __xgo_register_0(__xgo_func_info_0_0); ... }
```
This is inserted into the `trace.go` overlay/disk copy.

### Step 5: On go1.25, Blank Import Forces Compilation

`addBlankImports()` (`trace.go:515-583`) adds `import _ "github.com/xhd2015/xgo/runtime/trace"` to test source files. This blank import forces `runtime/trace` to be compiled and linked into the test binary. The instrumented `init()` runs, calling `XgoRegister(&TraceInfo)` which registers `Trace` in functab.

### Step 6: On go1.24, No Blank Import → No Compilation

On go1.24, `importRuntimeDepGenOverlay` never runs. No blank import. `runtime/trace` is never imported → never compiled → never linked. The instrumented `trace.go` in the overlay is irrelevant because nothing imports it. Trace's `init()` never runs.

### Step 7: Why `AllowInstrument=false` Doesn't Help

**`instrument/config/config.go:75-77`** — `CheckInstrument`:
```go
if isXgoRuntimePkg(pkgPath) {
    return true, false  // isXgo=true, allow=false
}
```

xgo runtime packages (`runtime/trace`, `runtime/mock`, etc.) have `AllowInstrument=false`. The main instrument loop skips them. But `registerFuncTab` does not — it's missing the `AllowInstrument` guard.

On go1.24 this doesn't matter because the generated code is never compiled (package not imported). On go1.25, the blank import forces compilation.

### Step 8: Init Order Difference

The test expects this order in `funcInfos`:
```
[0] SomeVar
[1] SomeVarWithoutType
[2] SomeFunc
[3] SomeType.ValueMethod
[4] (*SomeType).PtrMethod
[5] TestFuncTab
```

On go1.25 with rebuilt stdlib, the init DAG order puts `runtime/trace`'s init before the test package's init, so `Trace` ends up at index 0, shifting everything.

## Git History

| Commit | Description |
|--------|-------------|
| `55be783` | Add Go 1.25 support (version bounds, adapter files) |
| `3123289` | Avoid GOMODCACHE overlays — introduce `needLocalRuntime` |
| `c8a8e28` | Fix `forceLoad` parameter to also use `needLocalRuntime` |
| `58f7584` | Add Go 1.26 support (extends `needLocalRuntime` to `>= 25`) |

## Potential Fixes

### A. Add `AllowInstrument` Guard to `registerFuncTab`

```go
func registerFuncTab(packages *edit.Packages) {
    for _, pkg := range packages.Packages {
        if !pkg.AllowInstrument {
            continue
        }
        ...
    }
}
```

This prevents functab entries from being generated for ANY non-instrumentable package. It's the most consistent fix with the existing design (matches main loop behavior).

### B. Skip `registerFuncTab` for xgo runtime packages specifically

Target only packages where `pkg.Xgo` is true. Tighter scope but less general.

### C. Fix the test to filter by package path

Uncomment the filter at `func_tab_test.go:102-104`. Only fixes the symptom, not the root cause.

### D. Don't generate functab registration for packages traced via `ForceInPlace`

`runtime/trace.Trace` is trapped via `ForceInPlace: true` in `LinkXgoRuntime`. The functab entry isn't needed because xgo's trap mechanism works independently.
