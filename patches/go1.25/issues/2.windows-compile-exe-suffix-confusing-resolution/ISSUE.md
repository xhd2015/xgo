---
os: windows
version: go1.25
---

# Windows: Stale `.exe` Shadows New Binary After GOROOT Rebuild

## Summary

On Windows, when a directory contains both `compile` (no extension) and
`compile.exe`, the OS resolves `compile` to `compile.exe` first — `.exe` has
highest priority in Windows `PATHEXT`. xgo's instrumented GOROOT rebuild
produces a NEW compiler binary at `compile` (no `.exe`, because xgo's
`__config__.json` omitted the extension), but `syncGoroot` had already copied
the original `compile.exe` into the same directory. Windows resolution picks
the OLD `.exe` → compiler patches never run → functab empty.

**This is NOT a `go build` issue.** `go build -o ./binary` uses the path
literally on all platforms. The issue is Windows OS-level executable
resolution where a stale `.exe` file shadows a newer file with the same
base name.

### Problem chain

1. `syncGoroot` copies Go distribution → `compile.exe` in instrumented GOROOT
2. `__config__.json`: `-o .../compile` (no `.exe`)
3. `go build -o .../compile cmd/compile` → `compile` (new, instrumented)
4. Directory now has both: `compile` (NEW) + `compile.exe` (OLD)
5. Toolchain resolves `compile` → Windows picks `compile.exe` → OLD binary runs
6. `LinkXgoInit`, `__xgo_trap` → `XgoTrap` never execute
7. Result: `functab.GetFuncs()` returns 0 entries, all traps fail

### Discovery

CI log (`PR #385`, branch `dev-go1.25`): listing `pkg/tool/windows_amd64/`
revealed the smoking gun:

```
compile         size=29274624  ← NEW (instrumented, no .exe)
compile.exe     size=22063104  ← OLD (original, from syncGoroot)
```

Manual rebuild with `go build ... -o compile_test.exe` produced a different
md5, confirming the source WAS correctly patched — just the output path was
wrong.

## Reproduction

```bash
cd patches/go1.25/issues/windows-compile-exe-suffix-confusing-resolution/repro
go run main.go
```

### What it does

1. Builds `go build -o ./mybinary .` (program prints "hello")
2. Renames `mybinary` → `mybinary.exe` (simulating stale `.exe` from `syncGoroot`)
3. Modifies source to print "hello exe", rebuilds `-o ./mybinary` (creates new binary)
4. Invokes `mybinary` via OS shell

| Platform | OS resolution | Output |
|----------|--------------|--------|
| Windows | Resolves `mybinary` → `mybinary.exe` (stale) | `hello` |
| Linux/macOS | Runs `mybinary` directly (new) | `hello exe` |

### GitHub Actions

| Workflow | OS | Expected |
|----------|----|----------|
| `debug-exe-suffix-linux.yml` | ubuntu-latest | Pass (exit 0) |
| `debug-exe-suffix-windows.yml` | windows-latest | Fail (exit 1), `continue-on-error: true` |

## Investigation Flow

### 1. Overlay not applied — FALSE
Overlay source files DID contain `__xgo_trap_0` and `__xgo_func_info_0_0`
markers. Instrumentation was working correctly.

### 2. RegisterFuncTab not adding `__xgo_init_` — FALSE
Debug logging on Windows confirmed `VarDefStmts=2 VarRegStmts=2` —
registration code WAS being generated.

### 3. Compiler patches not applied to source — FALSE
Instrumented GOROOT source inspection confirmed:
- `src/cmd/compile/internal/gc/main.go` had the `xgo_rewrite_internal/patch` import
- `src/cmd/compile/internal/xgo_rewrite_internal/patch/` contained 2 patched files
- `.xgo.patch` files were correctly applied

### 4. Compiler not actually rebuilt — TRUE
md5 comparison before/after:

| Binary | Original MD5 | After xgo setup |
|--------|-------------|-----------------|
| `go.exe` | `A2B8F587...` | `A2B8F587...` (unchanged) |
| `compile.exe` | `BF344BF9...` | `BF344BF9...` (unchanged) |

### 5. Manual rebuild works — TRUE
Running the exact `go build` command manually produced a binary with DIFFERENT
md5 — proving source was correctly patched and build could succeed.

### 6. Path separator issues — PARTIAL
`__config__.json` used `/` in paths. On Windows, `${INSTRUMENT_GOROOT}` resolves
to backslash paths, creating mixed separators. Added path normalization in
`apply.go`.

### 7. Go build cache interference — POSSIBLE FACTOR
Subprocess inherited `GOCACHE`/`GOPATH` from parent xgo process. Added clearing
of these in the subprocess environment.

### 8. Root cause found — listing tool directory
```
compile         size=29274624 md5=9DFE372E...  ← NEW (no .exe)
compile_test.exe size=29274624 md5=9DFE372E...  ← manual rebuild
compile.exe     size=22063104 md5=BF344BF9...  ← OLD (original)
```

**`go build -a -o .../compile cmd/compile`** wrote to `compile` (no `.exe`).
The original `compile.exe` remained unchanged. The Go toolchain resolved
`compile` → `compile.exe` via Windows PATHEXT → used the old unpatched binary.

## Root Cause

Windows `PATHEXT` prioritizes `.EXE`. When resolving a bare executable name
like `compile`, Windows tries `compile.EXE` first, then `.COM`, `.BAT`, etc.

xgo's `syncGoroot` copies the original Go distribution (including `compile.exe`)
into the instrumented GOROOT. The `__config__.json` rebuilds the compiler with
`-o .../compile` (no `.exe`, path used literally by `go build`). The directory
now has both `compile` (new) and `compile.exe` (old). Windows prioritizes
`.exe` → stale binary runs → compiler patches never applied.

## Fix

**`instrument/patch/apply.go`:**
- `ensureOutputExeSuffixOnWindows()` — scans args for `-o <path>`, appends `.exe`
  if path lacks a file extension on Windows
- `ensureExeSuffixOnWindows()` — ensures binary path (first arg) has `.exe` on Windows
- Path normalization: convert `/` to `\` on Windows after `${VAR}` substitution
- Clear `GOCACHE` and `GOPATH` in subprocess environment

**`patches/go1.25/__config__.json`:**
- Added `cmd_windows` entries with `.exe` in output paths
- Added `-v` flag for verbose build output

## Appendix: Key Diagnostic Techniques

| # | Technique | What it reveals |
|---|-----------|----------------|
| 1 | md5 checks before/after | Whether rebuild actually happened |
| 2 | File listing of output directory | Mismatched filenames (`compile` vs `compile.exe`) |
| 3 | Source inspection in instrumented GOROOT | Whether patches/copies were applied |
| 4 | Debug logging in key functions | `RegisterFuncTab`, `instrumentFuncTrap` etc. |
| 5 | Manual rebuild test in CI | Isolate `go build` from xgo orchestration |

## Appendix: Debugging Escalation Path

```
Issue reported
    │
    ├─ --log-debug → functab/trap registration logs
    │   ├─ RegStmts empty → source instrumentation issue (overlay/load)
    │   └─ RegStmts present but functab empty → compiler rewrite issue
    │
    ├─ apply.go stderr → GOROOT patching/rebuild commands
    │   ├─ Commands not running → env/path issue
    │   └─ Commands run, binary unchanged → output path issue
    │
    ├─ Compiler file tracing → link rewrite execution
    │   ├─ PATCH not called → compiler binary not rebuilt
    │   └─ LookupRuntime NIL → runtime symbols missing
    │
    ├─ -v flag → verbose build output
    │
    └─ CI workflow → reproducible cross-platform diagnosis
```

## Related

- [windows-raw-string-syntax-rewrite-issue](../windows-raw-string-syntax-rewrite-issue/) — CRLF in raw string literals
- [windows-git-checkout-introduces-cr](../windows-git-checkout-introduces-cr/) — root cause of `\r` in source files
- [PR #385](https://github.com/xhd2015/xgo/pull/385) — original investigation PR
