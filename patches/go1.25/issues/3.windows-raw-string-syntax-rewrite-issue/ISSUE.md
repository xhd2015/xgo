---
os: windows
version: go1.25
---

# Raw String Constant Position Bug in Syntax Rewrite

## Summary

xgo instruments Go source files by inserting stub declarations and variable trap
functions into the AST. The insertion position is computed from
`ast.GenDecl.End()`, which for `ast.BasicLit` (string literals) returns
`ValuePos + len(Value)`. For declarations containing **raw string constants**
(backtick-delimited), this computed end position diverges from the actual byte
offset in the source, causing inserted code to land **inside** the raw string
literal. The Go parser then treats the inserted code as string content, making
stub declarations invisible — resulting in `undefined:` and `syntax error`
compilation failures.

The bug exists on all platforms but is only exposed on **Windows Go 1.25**,
where `goVersion.Minor >= 25` triggers `needLocalRuntime` (full instrumentation
for all packages), including xgo tooling code that happens to contain raw string
constants.

### Where `\r` Comes From (Root Cause)

The `\r` bytes that cause position divergence originate from **git** itself.
On Windows, Git for Windows defaults to `core.autocrlf=true`, which converts
LF (`\n`) to CRLF (`\r\n`) on **every** `git checkout` and `git clone`. This
conversion applies to ALL files — including Go source files with raw string
literals. The `\r` bytes land **inside** backtick-delimited strings.

```
Source authored (LF):          git clone on Windows (core.autocrlf=true):
const help = `\nUsage:...`     const help = `\r\nUsage:\r\n...`
```

The Go parser then:
1. Tracks byte positions including `\r` in `token.FileSet`
2. Strips `\r` from `BasicLit.Value` (the logical string value)
3. Computes `End() = ValuePos + len(Value)` — short by the `\r` count
4. Result: `genDecl.End()` points **inside** the raw string literal

xgo then inserts stubs at this wrong position, burying them in raw string content.

A standalone reproduction demonstrating this git behavior is at:
**[windows-git-checkout-introduces-cr](../windows-git-checkout-introduces-cr/)** —
including scripts that show before-clone vs after-clone hex dumps and CI
workflows confirming the behavior on Windows vs Linux.

### Affected error patterns

All errors were surfaced by running `go run ./script/run-test` with the
instrumented GOROOT:

#### Error 1 — `undefined: __xgo_trap_6` (Windows + Linux)

```bash
go run ./script/run-test --include go1.25.10 --install-xgo --with-setup --log-debug -v ./cmd/xgo/exec_tool
```

```
cmd/xgo/exec_tool/tool_exec_main.go:...: undefined: __xgo_trap_6
cmd/xgo/exec_tool/tool_exec_main.go:...: undefined: __xgo_func_info_6_0
FAIL    github.com/xhd2015/xgo/cmd/xgo/exec_tool [build failed]
```

Root declaration: `const ToolExecMainHelp = \`...\`` — stubs inserted inside the
raw string.

#### Error 2 — `misplaced go:embed directive` (Linux, after Iteration 1 fix)

```bash
go run ./script/run-test --include go1.25.10 --install-xgo --with-setup --log-debug -v ./cmd/trace
```

```
cmd/xgo/trace/render/resource.go:12: misplaced go:embed directive
FAIL    github.com/xhd2015/xgo/cmd/trace [build failed]
```

Root cause: `GetFuncInsertPosition` used `nextDecl.Pos()` unconditionally,
inserting code between `//go:embed script.js` and `var script string`.

#### Error 3 — `undefined: copyIconSVG_xgo_get` (Windows only, after Iteration 2 fix)

```bash
go run ./script/run-test --include go1.25.10 --install-xgo --with-setup --log-debug -v ./cmd/xgo/trace/render/
```

```
cmd\xgo\trace\render\render.go:176: undefined: copyIconSVG_xgo_get
cmd\xgo\trace\render\render.go:176: undefined: checkedIconSVG_xgo_get
FAIL    github.com/xhd2015/xgo/cmd/trace [build failed]
```

Root declaration: `var copyIconSVG = \`...\`` — `_xgo_get()` functions inserted
inside the raw string value by the variable trap path.

#### Error 4 — `syntax error: non-declaration statement` (after initial var trap fix)

```bash
go run ./script/run-test --include go1.25.10 --install-xgo --with-setup --log-debug -v ./cmd/xgo/trace/render/
```

```
cmd/xgo/trace/render/resource.go:21: syntax error: non-declaration statement outside function body
cmd/xgo/trace/render/resource.go:26: syntax error: non-declaration statement outside function body
FAIL    github.com/xhd2015/xgo/cmd/xgo/trace/render [build failed]
```

Root cause: `;` separator from variable trap insertion landed before the next
declaration when using `nextDecl.Pos()` as insertion point.

## Root Cause

The Go `go/ast` package computes `BasicLit.End()` as:

```
End() = ValuePos + len(Value)
```

For raw string constants, this position disagrees with the actual byte offset
in the source when `\r` bytes are present. This happens on Windows where
git's `core.autocrlf=true` injects `\r` into all files on checkout (see
**Where `\r` Comes From** above, and the standalone reproduction at
[windows-git-checkout-introduces-cr](../windows-git-checkout-introduces-cr/)).
The same divergence was first observed in `Append` (Root Cause #1 in the
handoff doc), which was fixed by bypassing AST positions entirely
(`len(edit.Buffer().Old())`).

The same class of bug existed in two other insertion paths:

1. **`GetFuncInsertPosition()`** (`instrument/patch/init.go`): Computes where
   `RegisterFuncTab` inserts package-level stubs (`__xgo_trap_N`,
   `__xgo_register_N`, `__xgo_func_info_N_X`). Uses the **first non-import
   GenDecl's `End()`** as the insertion point.

2. **`rewriteVarDefAndRefs()`** (`instrument/instrument_var/trap_var.go`): Inserts
   `_xgo_get()` / `_xgo_get_addr()` functions after each trapped variable
   declaration. Uses `decl.Decl.End()` for each **individual GenDecl**.

## Scope

| Path | What is inserted | Affected by raw string `End()`? |
|------|-----------------|--------------------------------|
| `instrument_reg.RegisterFuncTab` → `patch.AddCode` | Package stubs + func info vars | Yes — `GetFuncInsertPosition` uses `genDecl.End()` |
| `instrument_reg.RegisterFuncTab` → `patch.Append` | `__xgo_init_N()` function | Yes — **Fixed** (uses buffer length) |
| `instrument_var.rewriteVarDefAndRefs` | `varName_xgo_get()` functions | Yes — uses `decl.Decl.End()` |
| `patch/syntax.AfterFilesParsed` (compiler-side) | Autogen file via `addFile()` | No — generates `.go` source, not overlay edit |

## Fix Loop

The fix was developed through a CI iteration cycle:

```
1. CODE CHANGE → 2. COMMIT & PUSH → 3. FETCH CI LOGS → 4. DECIDE
```

### Iteration 1: Fix `GetFuncInsertPosition` (8a45f21)

Changed `GetFuncInsertPosition` to return `nextDecl.Pos()` instead of
`genDecl.End()` for the first non-import GenDecl. `AddCode` updated to use
`\n` prefix and suffix.

**Result**: Windows CI passes for `exec_tool`. But Linux CI reveals a
regression — `misplaced go:embed directive` in `cmd/xgo/trace/render`.

### Iteration 2: Guard against `go:` directives (d15df2f)

Added `HasGoDirective()` check: when the next declaration has a `//go:`
directive comment (`go:embed`, `go:linkname`, etc.), fall back to
`genDecl.End()` to avoid separating the directive from its declaration.

Extracted `GenDeclSafeEnd()` helper for reuse.

**Result**: Both `exec_tool` and `trace/render` pass locally on macOS.

### Iteration 3: Fix variable trap insertion (f2d0518)

Windows CI reveals a new failure: `undefined: copyIconSVG_xgo_get` in
`trace/render`. This is the **same root cause** in a different code path —
`rewriteVarDefAndRefs` uses `decl.Decl.End()` for the `copyIconSVG` and
`checkedIconSVG` variable declarations (which have raw string values).

Changes:
- `trap_var.go` uses `GenDeclSafeEnd()` via a new `safeGenDeclEnd()` wrapper
- Insertion changed from `;` separator to `\n` separator to work correctly
  with both `genDecl.End()` (after declaration) and `nextDecl.Pos()` (before
  next declaration)

**Result**: All three tests (`exec_tool`, `trace/render`, `instrument_raw_string`)
pass on both macOS and Windows CI.

## Fix Details

### `instrument/patch/init.go`

```go
// GenDeclSafeEnd returns a safe insertion point after a GenDecl.
// When the next declaration has no go: directive, uses nextDecl.Pos()
// (computed by the scanner, always correct). Otherwise falls back to
// genDecl.End() to avoid separating directives from their declarations.
func GenDeclSafeEnd(file *ast.File, declIndex int) token.Pos {
    if declIndex+1 < len(file.Decls) && !HasGoDirective(file.Decls[declIndex+1]) {
        return file.Decls[declIndex+1].Pos()
    }
    return file.Decls[declIndex].End()
}
```

### `instrument/instrument_var/trap_var.go`

```go
// safeGenDeclEnd wraps patch.GenDeclSafeEnd by finding the decl's index
func safeGenDeclEnd(file *edit.File, decl *edit.Decl) token.Pos {
    syntaxFile := file.File.Syntax
    for i, d := range syntaxFile.Decls {
        if d == decl.Decl {
            return patch.GenDeclSafeEnd(syntaxFile, i)
        }
    }
    return decl.Decl.End()
}
```

Insertion changed from `;` to `\n`:

```go
// Before:
fileEdit.Insert(end, ";")
fileEdit.Insert(end, code)

// After:
fileEdit.Insert(end, "\n"+code+"\n")
```

The `\n` separator works correctly with both position sources:
- `genDecl.End()`: `...var styles string\nfunc styles_xgo_get()...\n`
- `nextDecl.Pos()`: `...var copyIconSVG = \`...\`\nfunc copyIconSVG_xgo_get()...\nvar checkedIconSVG...`

## Key Files

### Core fix files (3 files in diff)

| File | What changed |
|------|--------------|
| `instrument/patch/init.go` | `GenDeclSafeEnd()` — returns safe position after GenDecl using `nextDecl.Pos()` when no `go:` directive; `HasGoDirective()` — checks for `//go:` doc comments; `GetFuncInsertPosition` uses `GenDeclSafeEnd` |
| `instrument/instrument_var/trap_var.go` | `safeGenDeclEnd()` — finds decl index and delegates to `patch.GenDeclSafeEnd`; insertion changed from `;` separator to `\n` separator to work with both `genDecl.End()` and `nextDecl.Pos()` |
| `runtime/test/build/instrument_raw_string/` | New integration test: `hello.go` with raw string const + function, `hello_test.go` verifying compilation succeeds |

The full core fix diff is at [`core_fix.diff`](core_fix.diff) (170 lines, 5 files).
A self-contained repro is at [`repro/repro_test.go`](repro/repro_test.go) — run with `go test -C repro -v`.

### Supporting files

| File | Change |
|------|--------|
| `.github/workflows/go-debug.yml` | Added `trace/render` test step |

## Verification

```bash
# Local (macOS):
go run ./script/run-test --include go1.25.10 --install-xgo --with-setup -v \
  ./cmd/xgo/exec_tool \
  ./cmd/xgo/trace/render \
  ./runtime/test/build/instrument_raw_string

# CI (Windows Go 1.25):
# https://github.com/xhd2015/xgo/actions/workflows/go-debug.yml
# Steps: Actual test (exec_tool) + Actual test (trace/render) → both PASS
```

## Related

- `HANDOFF_WINDOWS_EXEC_TOOL_FAILURE.md` — initial handoff with Root Cause #1 and #2
- `patches/go1.25/issues/runtime-lib-self-require/ISSUE.md` — another go1.25 compatibility issue
- `doc/development/WHY_GO_1_25_FAIL_ON_TestFuncNames.md`
