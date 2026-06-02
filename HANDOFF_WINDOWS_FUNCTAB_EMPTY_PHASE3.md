# HANDOFF: Windows Go 1.25 — Functab Empty (phase 3)

**Continuation of:** `HANDOFF_WINDOWS_FUNCTAB_EMPTY_PHASE2.md`
**Branch:** `dev-go1.25`
**Last commit:** `ca73a56`
**PR:** https://github.com/xhd2015/xgo/pull/385

## Goal

Fix Windows Go 1.25 CI where `functab.GetFuncs()` returns 0 entries. Linux/macOS pass.

## Updated Status (Narrowed Down)

### Theories resolved

| Theory | Status | Evidence |
|--------|--------|---------|
| C: Overlay not applied | **FALSE** | Overlay source HAS `__xgo_trap_0` and `__xgo_func_info_0_0` markers. The overlay IS applied. |
| B: `LookupRuntime` returns nil | **UNLIKELY** | Other files (e.g., `func_tab_test.go`) successfully get `__xgo_init_` markers on Windows, meaning `LookupRuntime` works. |
| A: `LinkXgoInit` never sees `__xgo_init_0` | **CONSEQUENCE, not cause** | `__xgo_init_` markers are never generated in the functab_mini overlay source. |

### New root cause narrowed down

**`__xgo_init_` markers are NOT present in the functab_mini overlay file on Windows**, though `__xgo_trap_0` IS present. On Linux, both are present. This means the functab registration code (`registerFuncTab` → `GetFileRegStmts`) is not adding `__xgo_init_` for functab_mini on Windows.

Specifically, the `instrument_reg.RegisterFuncTab` function in `instrument/instrument_reg/instrument_reg.go` returns early at line 32 when `res.VarDefStmts` is empty:

```go
res := compiler_extra.GetFileRegStmts(fileDecls, stdlib, FILE_VAR, FILE_VAR_FOR_VAR, perFilePkgNames)
if len(res.VarDefStmts) == 0 {
    return  // <-- __xgo_init_ and __xgo_register_ are NOT generated
}
```

`fileDecls` is built from `file.TrapFuncs`, which should contain `Hello`. On Linux it does (generating `__xgo_init_`), on Windows it apparently doesn't (or `GetFileRegStmts` returns empty despite populated `fileDecls`).

### Side-by-side comparison (latest CI)

| Metric | Linux | Windows |
|--------|-------|---------|
| Trap interceptor fires | ✅ Yes | ❌ No |
| functab.GetFuncs() | 624 entries | 0 entries |
| `__xgo_trap_0` in overlay | ✅ | ✅ |
| `__xgo_init_` in functab_mini overlay | ✅ | ❌ (other files DO have it) |
| `LinkXgoInit` debug log | ✅ Has entries | ❌ Empty (compiler not calling our code OR file write fails) |

## Changes Made (Keep These)

### 1. `RebaseAbsPath` fix — `support/fileutil/path.go:40`

Fixed drive letter handling. Before: `D:\a\...\go.mod` → `overlay\D\a\...\go.mod`. After: `overlay\a\...\go.mod`.

```go
func RebaseAbsPath(root string, absPath string) string {
    if runtime.GOOS != "windows" {
        return filepath.Join(root, absPath)
    }
    idx := strings.Index(absPath, ":")
    if idx < 0 {
        return filepath.Join(root, absPath)
    }
    rest := absPath[idx+1:]
    if len(rest) > 0 && (rest[0] == '/' || rest[0] == '\\') {
        rest = rest[1:]
    }
    return filepath.Join(root, rest)
}
```

### 2. Trap interceptor test — `runtime/test/functab/functab_mini/functab_mini_test.go`

Added trap interceptor check (`trap.AddInterceptor`) to test whether overlay trap wrappers are applied. Imports `context`, `core`, `trap`. This is the key diagnostic for Theory C.

### 3. CI workflows

| File | Purpose |
|------|---------|
| `.github/workflows/go-debug.yml` | Windows Go 1.25 debug workflow (targets functab_mini, has diagnostics) |
| `.github/workflows/go-debug-linux.yml` | Linux mirror for side-by-side comparison |

Both use `continue-on-error: true` on the main test step (so diagnostics run even on failure). Both trigger on `pull_request: master` only (push-to-dev-go1.25 removed per instruction).

### 4. Compiler patch tracing (partial) — `patch/patch.go` + `patch/link/link_ir.go`

Added `appendLog`/`appendDebug` functions that write to `os.TempDir()/xgo_link_init_debug.log`. On Linux this works and shows `LinkXgoInit` being called. On Windows the log file is always empty — either the compiler isn't rebuilt with our code, or file write fails.

Also synced copies at `cmd/xgo/asset/compiler_patch_gen/patch.go` and `cmd/xgo/asset/compiler_patch_gen/link/link_ir.go` (must match — they're embedded into xgo binary at build time).

## What Didn't Work

1. **stderr-based compiler logging** — Go suppresses compiler stderr on successful builds. `fmt.Fprintf(os.Stderr, ...)` calls from `Patch()`/`LinkXgoInit` never appear in logs.

2. **File-based compiler logging on Windows** — `os.OpenFile(os.TempDir()/..., ...)` produces an empty file. Could be:
   - Silent `OpenFile` failure (our error handling just returns)
   - Compiler binary not actually rebuilt with our code (Go build cache)
   - Different `TEMP` directory than what the CI dump step checks
   - **Recommendation:** skip further compiler tracing attempts; focus on the overlay generation side.

3. **`continue-on-error: true` masks failures** — the workflow shows "success" even when the test fails. Remove this when closing in on a fix.

## Next Steps

### Priority 1: Investigate why `registerFuncTab` doesn't add `__xgo_init_` on Windows

The core issue is in the xgo instrumentation code (not compiler patches). Key code path:

```
cmd/xgo/instrument.go:348 registerFuncTab()
  └─ instrument/instrument_reg/instrument_reg.go:13 RegisterFuncTab()
       └─ buildCompilerExtra() — uses file.TrapFuncs (from instrumentFuncTrap)
       └─ compiler_extra.GetFileRegStmts() — returns VarDefStmts
            └─ if empty → return (no __xgo_init_ generated)
```

**Investigation approaches:**

a) **Add file-based logging to `RegisterFuncTab`** (in `instrument/instrument_reg/instrument_reg.go`):
   - Log `len(file.TrapFuncs)`, `len(res.VarDefStmts)`, and `pkgPath`
   - This runs in the xgo binary (not the Go compiler), so file/log output is visible
   - Compare Linux vs Windows output

b) **Check `TrapFuncs` population in `instrumentFuncTrap`** (`cmd/xgo/instrument.go:246`):
   - For `functab_mini_test.go`, verify `file.TrapFuncs` contains `Hello`
   - Check if the test-file skip condition (line 247: `!pkg.Main && _test.go`) applies

c) **Check `GetFileRegStmts` logic** — does it have Windows-specific path comparisons that fail?

d) **Run with `--log-debug`** and grep for wrapper-related messages. Add more debug logs to the instrumentation code if needed.

### Priority 2: Verify compiler patch is applied

Confirm the patched `cmd/compile` binary is actually rebuilt with our changes:

```yaml
- name: Check compiler binary mtime
  run: Get-Item "$HOME\.xgo\go-instrument\*\go1.25.*\pkg\tool\windows_amd64\compile.exe" | Select-Object FullName, LastWriteTime, Length
```

If the compiler IS rebuilt but `appendLog` still doesn't write, remove the debug code from patch files.

### Priority 3: Fix the `-modfile` path (partially done)

The `RebaseAbsPath` fix resolved the `D` folder issue. But observe whether Go accepts the `-modfile` path with backslashes. From the command:
```
go test -modfile D:\...\overlay\a\xgo\xgo\runtime\test\go.mod
```

Check if Go correctly resolves this path on Windows.

### CI iteration

```bash
# 1. Make changes, commit, push
git add -A && git commit -m "debug: <describe>" && git push origin dev-go1.25

# 2. Wait and fetch logs
sleep 240 && gh run view $(gh run list --repo xhd2015/xgo --workflow "Go Debug" --limit 1 --json databaseId --jq '.[0].databaseId') --log | grep -E "(TRAP|functab|__xgo_init|register)" | head -20

# 3. Also check Linux for comparison
gh run view $(gh run list --repo xhd2015/xgo --workflow "Go Debug Linux" --limit 1 --json databaseId --jq '.[0].databaseId') --log | grep -E "(TRAP|functab)" | head -10
```

## Key Files Reference

| File | Purpose |
|------|---------|
| `cmd/xgo/instrument.go:348` | `registerFuncTab` — entry point for functab registration code gen |
| `instrument/instrument_reg/instrument_reg.go:13` | `RegisterFuncTab` — per-file functab registration, returns early if `VarDefStmts` empty |
| `instrument/instrument_reg/instrument_reg.go:32` | **The skip condition** — `if len(res.VarDefStmts) == 0 { return }` |
| `cmd/xgo/instrument.go:246-258` | `instrumentFuncTrap` loop — populates `file.TrapFuncs` |
| `cmd/xgo/instrument.go:247-249` | Test file skip: `if !pkg.Main && _test.go { continue }` |
| `support/fileutil/path.go:40` | `RebaseAbsPath` — **fixed** drive letter handling |
| `runtime/test/functab/functab_mini/functab_mini_test.go` | Test file with trap interceptor check |
| `patch/link/link_ir.go` | `LinkXgoInit` — compiler rewrite (has debug `appendDebug` function) |
| `patch/patch.go` | `Patch()` — compiler entry (has debug `appendLog` function) |
| `cmd/xgo/asset/compiler_patch_gen/` | **Must be synced** when `patch/` changes (embedded into xgo binary) |
| `patches/go1.25/__config__.json` | Rebuilds patched compiler & stdlib |
| `.github/workflows/go-debug.yml` | Windows diagnostic workflow |
| `.github/workflows/go-debug-linux.yml` | Linux comparison workflow |

## Cleanup Needed (after fix is confirmed)

- Remove `appendLog`/`appendDebug` from `patch/patch.go` and `patch/link/link_ir.go`
- Sync `cmd/xgo/asset/compiler_patch_gen/` copies
- Remove trap interceptor test from `functab_mini_test.go` (or keep as validation)
- Remove `continue-on-error: true` from go-debug workflows (or remove the workflows entirely)
- Consider deleting `go-debug-linux.yml` if no longer needed
