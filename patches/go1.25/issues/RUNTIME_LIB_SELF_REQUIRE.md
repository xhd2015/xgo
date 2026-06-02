# Self-Require Issue: Runtime Module Requiring Itself

## Summary

Go 1.25 forbids `-overlay` from replacing files under `GOMODCACHE`. To work around this, xgo's `needLocalRuntime` path (`cmd/xgo/main.go:710`, triggered when `goVersion.Minor >= 25`) copies the runtime to a local directory and uses `-modfile` with a `replace` directive to redirect the runtime module to the local copy.

To make the `replace` directive syntactically valid, a matching `require` must be added to the `go.mod`. However, when the **project being tested IS `github.com/xhd2015/xgo/runtime` itself** (e.g., xgo's own CI running `go test ./core/...`), this creates a **self-require** — the module requires itself. Go 1.25's stricter module validation detects that no source code actually imports the module and rejects it:

```
go: updates to go.mod needed; to update it:
	go mod tidy
```

## Root Cause

`loadDependency()` at `cmd/xgo/trace.go:383-390` unconditionally injects:

```
go mod edit
  -require=github.com/xhd2015/xgo/runtime@v<version>
  -replace=github.com/xhd2015/xgo/runtime=<local_copy>
  <project's go.mod copy>
```

- For a **user project** (`github.com/my/app`): the `require` is a legitimate external dependency → Go accepts it.
- For **xgo/runtime itself**: it's a self-require → Go 1.25 rejects it.

## Scope

| Scenario | Go 1.24 | Go 1.25 |
|----------|---------|---------|
| User project (`go test ./...`) | PASS (overlay only, no -modfile) | PASS (self-require is for xgo/runtime, not the project) |
| xgo/runtime CI (`go test ./core/...`) | PASS (overlay only, no -modfile) | **FAIL** (self-require injected) |
| `xgo test ./core/...` | PASS (xgo's own test command) | PASS (xgo patches code before compiling) |

## Fix

Two bypasses are needed when the main module IS `github.com/xhd2015/xgo/runtime`:

### 1. Skip self-require (modfile)

In `loadDependency()`, skip the `require` + `replace` because the runtime code is already the project itself.

```go
// cmd/xgo/trace.go:386
if mainModule != constants.RUNTIME_MODULE {
    go mod edit -require=RUNTIME_MODULE@VERSION -replace=RUNTIME_MODULE=<local_copy>
}
```

### 2. Skip blank import injection (import cycle)

`addBlankImports()` copies `core/func.go` and injects `import _ "github.com/xhd2015/xgo/runtime/trace"` to auto-load tracing. For xgo/runtime itself, this creates an import cycle:

```
core/func.go  ──(overlay: import _ "runtime/trace")──→  trace/trace.go
                                                                │
                                                        imports internal/trap
                                                                │
                                                        internal/trap imports core
                                                                │
                                                    ←──── cycle ←────
```

Fix: skip `addBlankImports` entirely when building the runtime module.

```go
// cmd/xgo/trace.go:142
if mainModule != constants.RUNTIME_MODULE {
    fileReplace, err = addBlankImports(...)
}
```

## Key Files

| File | Role |
|------|------|
| `cmd/xgo/trace.go:383-390` | Injects self-require via `go mod edit` |
| `cmd/xgo/main.go:710` | `needLocalRuntime` gate (`goVersion.Minor >= 25`) |
| `instrument/constants/constants.go:5` | `RUNTIME_MODULE = "github.com/xhd2015/xgo/runtime"` |

## Verification

```bash
# Instrumented Go 1.25 binary — currently FAILS (self-require)
GOROOT=/Users/xhd2015/.xgo/go-instrument/go1.25.10_.../go1.25.10 \
  <goroot>/bin/go test -C /path/to/xgo/runtime ./core/...

# After fix: should PASS (self-require skipped for runtime module itself)
```

## Related

- `doc/development/OVERLAY_RESTRICTION_IN_GO_1_25_V.S._GO_1_24.md`
- `HANDOFF_GO_1_25_MODFILE_SELF_REQUIRE_FIX.md`
