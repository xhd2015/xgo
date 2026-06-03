# How to Adjust Code for Debugging xgo Issues

Practical recipes for re-adding debugging instrumentation. Add minimally, remove after diagnosis.

See [Go-1.25-Windows-Suffix-Issue.md](./Go-1.25-Windows-Suffix-Issue.md) for the full investigation flow and root cause analysis.

## Quick Reference

| Issue Type | Layer | What to Add | File(s) |
|---|---|---|---|
| GOROOT patching not executing | Layer 2 | stderr tracing | `instrument/patch/apply.go` |
| Compiler link rewrite not running | Layer 3 | file-based tracing | `patch/patch.go`, `patch/link/link_ir.go` |
| Functab/instrument tracing | Layer 1 | `--log-debug` flag | (CLI flag, no code change) |
| Build output unclear | Layer 4 | `-v` flag | `patches/go*/__config__.json` |
| Cross-platform CI issue | Layer 5 | Debug workflow | `.github/workflows/` + CI steps |

**After diagnosing:** revert all changes except Layer 1 (the `--log-debug` calls in `instrument_reg.go` and `cmd/xgo/instrument.go` remain in the codebase as opt-in diagnostics).

---

## Layer 1: `--log-debug` Flag (Built-in, Already Available)

**No code changes needed.** The codebase already has extensive `config.LogDebug` / `logDebug` calls behind the `--log-debug` flag. Pass it when running xgo:

```bash
xgo test --log-debug -v ./runtime/test/functab/functab_mini
```

Or set the env var:

```bash
XGO_LOG_DEBUG=stderr xgo test ./your/test
```

`--log-debug` accepts values: `stderr` (default), `stdout`, or a file path.

### What this reveals

The existing log calls trace:

| File | Key log lines |
|------|--------------|
| `cmd/xgo/instrument.go` | `registerFuncTab: total packages=... total files=...` |
| `cmd/xgo/instrument.go` | `instrumentFuncTrap: checking file=... trapped=... extra=...` |
| `cmd/xgo/instrument.go` | `instrumentFuncTrap: SKIP test file outside main: ...` |
| `instrument/instrument_reg/instrument_reg.go` | `RegisterFuncTab: file=... pkg=... trapFuncs=... trapVars=...` |
| `instrument/instrument_reg/instrument_reg.go` | `RegisterFuncTab: SKIP file=... (VarDefStmts empty, ...)` |
| `instrument/instrument_reg/instrument_reg.go` | `RegisterFuncTab: GENERATE file=... VarDefStmts=... VarRegStmts=...` |

### When to use

First pass for any functab, instrumentation, or trap-registration issue. If this already reveals the problem, you don't need the layers below.

---

## Layer 2: Stderr Tracing in `apply.go`

**File:** `instrument/patch/apply.go`

Add `fmt.Fprintf(os.Stderr, ...)` calls to trace GOROOT patching from config loading through command execution. This reveals whether patches are applied, what commands are run, and what environment is set.

### Step 1: Trace entry and config

Add after `func ApplyPatches(...) error {`:

```go
fmt.Fprintf(os.Stderr, "ApplyPatches: patchDir=%s goroot=%s xgoRepoRoot=%s\n", patchDir, goroot, xgoRepoRoot)
```

Add inside a loop after `cfg, err := LoadConfig(patchDir)`, before iterating `cfg.Copy`:

```go
for k, v := range extraEnv {
    fmt.Fprintf(os.Stderr, "ApplyPatches: extraEnv[%s]=%s\n", k, v)
}
```

### Step 2: Trace generate commands

After `args = substituteEnvSlice(args, extraEnv)`, add:

```go
fmt.Fprintf(os.Stderr, "ApplyPatches: effective args=%v (len=%d)\n", args, len(args))
```

After `cmdDir := resolveCwd(gen.Cwd, goroot, extraEnv)`, add:

```go
fmt.Fprintf(os.Stderr, "ApplyPatches: cmdDir=%s\n", cmdDir)
fmt.Fprintf(os.Stderr, "ApplyPatches: binary=%s args=%v\n", args[0], args[1:])
```

### Step 3: Trace environment

After `newEnv = append(newEnv, "GOROOT="+goroot, ...)`, add:

```go
fmt.Fprintf(os.Stderr, "ApplyPatches: env GOROOT=%s GOOS= GOARCH= GOPATH= GOCACHE=\n", goroot)
```

Inside the `for k, v := range gen.Env` loop, add:

```go
fmt.Fprintf(os.Stderr, "ApplyPatches: env %s=%s\n", k, v)
```

### What this reveals

| Log | Reveals |
|-----|---------|
| `patchDir=... goroot=...` | Confirms input paths are correct |
| `extraEnv[...]=...` | Shows `${VAR}` substitution values |
| `effective args=...` | Shows final command args after substitution |
| `cmdDir=... binary=...` | Confirms working directory and binary path |
| `env GOROOT=... GOOS=... GOPATH=...` | Confirms subprocess environment cleanup |

### When to use

When the GOROOT instrumentation appears not to run, or when `go build` commands in the generate steps fail or produce unexpected output.

---

## Layer 3: File-based Tracing in Compiler Patch Code

**Files:** `patch/patch.go`, `patch/link/link_ir.go`

These are the source files that become embedded in the Go compiler via `cmd/xgo/asset/compiler_patch_gen/`. When the compiler runs, `Patch()` is called to rewrite `__xgo_trap_` → `XgoTrap` etc. Tracing here reveals whether the compiler's rewrite code executes and what it processes.

### Step 1: Add helper function

Add this function to **both** `patch/patch.go` and `patch/link/link_ir.go`:

```go
func appendDebug(format string, args ...interface{}) {
	logPath := filepath.Join(os.TempDir(), "xgo_link_init_debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "XGO_LOG_ERROR: cannot open %s: %v\n", logPath, err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, format+"\n", args...)
}
```

Also add the needed imports:

```go
import (
	// ... existing imports ...
	"fmt"
	"os"
	"path/filepath"
)
```

### Step 2: Trace `Patch()` entry/exit (`patch/patch.go`)

```go
func Patch() {
	appendLog("PATCH called")
	linkFuncs()
	appendLog("PATCH done")
}

func linkFuncs() {
	count := 0
	funcs.ForEach(func(fn *ir.Func) bool {
		link.LinkXgoInit(fn)
		count++
		return true
	})
	appendLog("PATCH linkFuncs iterated %d funcs", count)
}
```

### Step 3: Trace `LinkXgoInit` and `linkStmts` (`patch/link/link_ir.go`)

In `LinkXgoInit`, after confirming `fnName` starts with `__xgo_init_`:

```go
appendDebug("LINK_INIT fn=%s body_stmts=%d", fnName, len(fn.Body))
```

In `linkStmts`, after finding a link:

```go
appendDebug("LINK_STMT sym=%s link=%s", symName, link)
```

If `LookupRuntime` returns nil:

```go
sym := typecheck.LookupRuntime(link)
if sym == nil {
	appendDebug("LINK_STMT LookupRuntime(%s) returned NIL", link)
	continue
}
```

### Step 4: Sync asset copies

After modifying `patch/patch.go` or `patch/link/link_ir.go`, sync the embedded copies:

```bash
go run ./script/generate cmd/xgo/asset/compiler_patch_gen
```

### Where the log appears

The log is written to `%TEMP%/xgo_link_init_debug.log` (e.g., `/tmp/xgo_link_init_debug.log` on Linux/macOS, `%TEMP%\xgo_link_init_debug.log` on Windows). On Linux/macOS:

```bash
cat /tmp/xgo_link_init_debug.log
```

On Windows (PowerShell):

```powershell
Get-Content "$env:TEMP\xgo_link_init_debug.log"
```

### What this reveals

| Log | Reveals |
|-----|---------|
| `PATCH called` / `PATCH done` | Confirms `Patch()` actually executed |
| `PATCH linkFuncs iterated N funcs` | Number of functions processed by the rewrite |
| `LINK_INIT fn=... body_stmts=N` | Which init function is being linked, how many statements |
| `LINK_STMT sym=... link=...` | Each `__xgo_trap_*` assignment being rewritten |
| `LINK_STMT LookupRuntime(...) NIL` | **Critical:** runtime symbol not found — rewrite cannot complete |

### When to use

When functab is empty despite correct source instrumentation. If `LookupRuntime` returns NIL for all symbols, the runtime package wasn't compiled with the right symbols. If the log file is empty or missing, the compiler rewrite (`Patch()`) never ran — the compiler binary is stale.

---

## Layer 4: Verbose Build Output in `__config__.json`

**File:** `patches/go1.25/__config__.json` (and corresponding files for other Go versions)

Add `-v` to build command args to see which packages `go build -a` actually compiles. For cmd array format:

```json
{
  "kind": "rebuild-compiler",
  "cmd": [
    "${INSTRUMENT_GOROOT}/bin/go",
    "build", "-v", "-a",
    "-o", "${INSTRUMENT_GOROOT}/pkg/tool/${GOOS}_${GOARCH}/compile",
    "cmd/compile"
  ],
  "cmd_windows": [
    "${INSTRUMENT_GOROOT}/bin/go.exe",
    "build", "-v", "-a",
    "-o", "${INSTRUMENT_GOROOT}/pkg/tool/${GOOS}_${GOARCH}/compile.exe",
    "cmd/compile"
  ]
}
```

Repeat for `rebuild-go`.

Then sync patches:

```bash
go run ./script/generate cmd/xgo/asset/patches
```

### What this reveals

Shows whether packages are being compiled or skipped (cached). If `go build -a` still produces "cached" output, the binary wasn't actually rebuilt despite `-a`.

---

## Layer 5: CI Diagnostic Workflow

**Reference files:** [workflow-go-debug.yml](./workflow-go-debug.yml), [workflow-go-debug-linux.yml](./workflow-go-debug-linux.yml)

Create a debug CI workflow for cross-platform diagnosis. Key diagnostic steps (see archived files for full PowerShell/bash):

### Step A: MD5 of original binaries

Record md5 of `go.exe` and `compile.exe` before xgo runs.

### Step B: Run xgo with debug output

```yaml
- run: go run ./script/run-test --install-xgo --with-setup --reset-instrument --log-debug -v ./runtime/test/functab/functab_mini 2>&1 | Tee-Object -FilePath xgo_debug_output.txt
  continue-on-error: true
```

### Step C: MD5 of instrumented binaries

Compare md5 values before/after to confirm rebuild happened.

### Step D: List files in output directory

```powershell
Get-ChildItem "$goroot\pkg\tool\windows_amd64" | ForEach-Object {
    $md5 = (Get-FileHash -Algorithm MD5 -Path $_.FullName).Hash
    Write-Output "$($_.Name) size=$($_.Length) md5=$md5"
}
```

Look for mismatched filenames (e.g., `compile` vs `compile.exe`).

### Step E: Manual rebuild verification

Reproduce the exact `go build` command that xgo should have run, to isolate xgo orchestration from the actual build:

```powershell
$env:GOROOT = $goroot
$env:GOTOOLCHAIN = "local"
$env:GO_BYPASS_XGO = "true"
& "$goroot\bin\go.exe" build -v -a -o compile_test.exe cmd/compile
```

### Step F: Source inspection in instrumented GOROOT

Verify patches were applied:

```powershell
# Check patched source compiles
Get-Content "$goroot\src\cmd\compile\internal\gc\main.go" -TotalCount 20
# Check patch directory
Get-ChildItem "$goroot\src\cmd\compile\internal\xgo_rewrite_internal\patch"
```

### When to use

When an issue only reproduces on a specific platform (e.g., Windows but not Linux). The debug workflow archives the diagnostic steps so you can re-run them on any branch.

---

## Summary: Debugging Escalation Path

```
Issue reported
    │
    ├─ Layer 1: --log-debug → reviews functab/trap registration logs
    │   │
    │   └─ RegStmts empty? → source instrumentation issue (overlay/load)
    │       RegStmts present but functab empty? → compiler rewrite issue
    │
    ├─ Layer 2: apply.go stderr → checks GOROOT patching/rebuild commands
    │   │
    │   └─ Commands not running? → env/path issue
    │       Commands run, binary unchanged? → output path issue
    │
    ├─ Layer 3: compiler file tracing → checks link rewrite execution
    │   │
    │   └─ PATCH not called? → compiler binary not rebuilt
    │       LookupRuntime NIL? → runtime symbols missing
    │
    ├─ Layer 4: -v flag → verbose build output
    │
    └─ Layer 5: CI workflow → reproducible cross-platform diagnosis
```
