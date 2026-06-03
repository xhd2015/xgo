# Debugging: Windows Go 1.25 — Functab Empty

**Date:** 2026-06-03
**PR:** https://github.com/xhd2015/xgo/pull/385
**Branch:** `dev-go1.25`

## Problem

On Windows Go 1.25 CI, `functab.GetFuncs()` returns 0 entries. Trap interceptors never fire. Linux and macOS pass with 634 entries.

## Investigation Flow

### 1. Overlay not applied (Theory C) — FALSE

The overlay source files DID contain `__xgo_trap_0` and `__xgo_func_info_0_0` markers. The overlay mechanism was working correctly.

### 2. RegisterFuncTab not adding `__xgo_init_` — FALSE

Added debug logging to `RegisterFuncTab()` in `instrument/instrument_reg/instrument_reg.go`. On Windows, the output confirmed:
```
RegisterFuncTab: file=.../functab_mini_test.go trapFuncs=2 trapVars=0
RegisterFuncTab: GENERATE file=... VarDefStmts=2 VarRegStmts=2
```
The source code instrumentation was complete. `__xgo_init_` markers WERE being generated.

### 3. Compiler patches not applied to source — FALSE

Directly inspected the instrumented GOROOT source via CI diagnostic steps:
- `src/cmd/compile/internal/gc/main.go` DID contain the `xgo_rewrite_internal/patch` import
- `src/cmd/compile/internal/xgo_rewrite_internal/patch/` directory DID contain the patched source files (2 files)
- `.xgo.patch` files were correctly applied by `applyPatchFiles`

### 4. Compiler not actually rebuilt — TRUE

Used md5 checks before and after the xgo setup. All md5 values matched the originals:

| Binary | Original MD5 | Instrumented MD5 |
|--------|-------------|-------------------|
| go.exe | `A2B8F58705EFF6BEC1F8BC9DB0C05F9F` | `A2B8F58705EFF6BEC1F8BC9DB0C05F9F` |
| compile.exe | `BF344BF90F2FB82F230555A00CD90432` | `BF344BF90F2FB82F230555A00CD90432` |

### 5. Manual rebuild works — TRUE

Added a manual rebuild step to the CI workflow that rebuilds compile.exe using the instrumented GOROOT. The manually built binary had a DIFFERENT md5:

```
NEW compile_test.exe MD5: 9DFE372EADD1E60896C1EF13899DBB97
```

This proved the source WAS correctly patched and the build COULD succeed — something was different about how xgo invoked the build.

### 6. Path separator issues — PARTIAL

The `__config__.json` uses literal `/` in paths like `${INSTRUMENT_GOROOT}/bin/go`. On Windows, `${INSTRUMENT_GOROOT}` resolves to a backslash path (from `filepath.Join`), creating mixed separators like `C:\...\go1.25.10/bin/go`. While Go's `os/exec` handles this, it's fragile. Added path normalization in `apply.go` to convert `/` to `\` on Windows.

### 7. Go build cache interference — POSSIBLE FACTOR

The subprocess inherits `GOCACHE` and `GOPATH` from the parent xgo process, which might cause `go build` to use stale cached objects. Added clearing of `GOCACHE` and `GOPATH` in the subprocess environment.

## Root Cause

Listed files in `pkg/tool/windows_amd64/` on the CI and found:

```
compile     size=29274624 md5=9DFE372EADD1E60896C1EF13899DBB97  ← NEW (correct)
compile_test.exe size=29274624 md5=9DFE372EADD1E60896C1EF13899DBB97  ← manual rebuild
compile.exe size=22063104 md5=BF344BF90F2FB82F230555A00CD90432  ← OLD (original)
```

**`go build -a -o .../pkg/tool/windows_amd64/compile cmd/compile` writes the output to a file literally named `compile` (no `.exe`).** The original `compile.exe` (copied by `syncGoroot`) remained unchanged. The Go toolchain looks for `compile.exe` and found the OLD, unpatched binary. The link rewrite (`LinkXgoInit`, `__xgo_trap` → `XgoTrap`) never ran, so functab was empty.

Go's `cmd/go` does NOT append `.exe` to the `-o` output path on Windows — it uses the path exactly as specified.

## Fixes Applied

### 1. Path normalization in `apply.go`

After `${VAR}` substitution, convert `/` to `filepath.Separator` on Windows to avoid mixed-separator paths.

### 2. `.exe` suffix on `-o` output paths

Added `ensureOutputExeSuffixOnWindows()` that scans command args for `-o <path>` and appends `.exe` if the path lacks a file extension on Windows.

### 3. Binary path `.exe` suffix

Added `ensureExeSuffixOnWindows()` to ensure the binary path (first arg) has `.exe` extension on Windows. While Go's `os/exec` does attempt `.exe` resolution via `lookExtensions`, making it explicit is safer.

### 4. Clear `GOCACHE` and `GOPATH`

The subprocess environment clears `GOCACHE` and `GOPATH` inherited from the parent to prevent stale cache interference during `go build -a`.

### 5. Verbose build output

Added `-v` flag to `rebuild-compiler` and `rebuild-go` in `__config__.json` for better build diagnostics.

## Key Diagnostic Techniques

1. **md5 checks before/after** — compare original and instrumented binaries to confirm whether rebuild actually happened
2. **File listing of output directory** — list all files with sizes and md5 to find mismatched filenames (e.g., `compile` vs `compile.exe`)
3. **Source inspection in instrumented GOROOT** — verify patches and copies were applied correctly
4. **Debug logging in key functions** — trace `RegisterFuncTab`, `instrumentFuncTrap`, etc.
5. **Manual rebuild test in CI** — reproduce the exact `go build` command to isolate the issue from the xgo orchestration code

## Files Changed

| File | Change |
|------|--------|
| `instrument/patch/apply.go` | `StringSliceOrSlice`, `EffectiveCmd()`, `substituteEnvSlice()`, env cleanup, `.exe` handling |
| `patches/go1.25/__config__.json` | Added `-v` to build commands |
| `instrument/instrument_reg/instrument_reg.go` | Debug logging (can be removed) |
| `cmd/xgo/instrument.go` | Debug logging (can be removed) |
| `.github/workflows/go-debug.yml` | Diagnostic CI steps |
