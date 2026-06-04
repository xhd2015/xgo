---
os: windows
version: all
---

# Git `core.autocrlf` Introduces `\r` into Raw String Literals on Clone

## Summary

Git on Windows defaults to `core.autocrlf=true`, which automatically converts
LF line endings (`\n`) to CRLF (`\r\n`) on `git checkout` / `git clone`. For
Go source files containing backtick-delimited **raw string literals**, this
conversion injects `\r` bytes (0x0d) **inside** the raw string content.

These extra `\r` bytes cause the Go parser's `ast.BasicLit.End()` to compute
incorrect positions, because `len(BasicLit.Value)` excludes `\r` while the
`token.FileSet` byte offsets include them. This directly causes xgo's
instrumentation to insert stub declarations at wrong byte offsets — landing
**inside** raw string literals instead of after them.

This is the **root cause** of the Windows-specific "undefined" compilation
errors documented in the sibling issue.

### Related issue

- [windows-raw-string-syntax-rewrite-issue](../windows-raw-string-syntax-rewrite-issue/) — the downstream xgo bug caused by this CRLF injection

## Reproduction

### Prerequisites

- Windows with `core.autocrlf=true` (Git for Windows default), OR
- Any platform with `git config --global core.autocrlf true`

### Steps

```bash
cd patches/go1.25/issues/windows-git-checkout-introduces-cr/repro
go run main.go
```

### Expected behavior

On Linux (where `core.autocrlf` is `false` or `input` by default):
```
RESULT: No \r found — file is pure LF. Safe.
```
Exit code 0.

On Windows (where `core.autocrlf=true` by default):
```
RESULT: \r FOUND — git autocrlf altered the source file.
        Raw string literals now contain \r bytes.
```
Exit code 1.

### GitHub Actions behavior

| Workflow | OS | Default autocrlf | Expected |
|----------|----|------------------|----------|
| `debug-git-crlf-checkout-linux.yml` | ubuntu-latest | false | Pass (exit 0) |
| `debug-git-crlf-checkout-windows.yml` | windows-latest | true | Fail (exit 1) |

The Windows workflow uses `continue-on-error: true` to allow verification
without failing the CI job.

## Mechanism

1. A Go source file is authored with LF-only line endings and a raw string literal:
   ```go
   const help = `
   Usage:
     -func match
   `
   ```

2. The file is committed to git with `core.autocrlf=false` (LF-only blob).

3. On `git clone` with `core.autocrlf=true`, git rewrites every `\n` to `\r\n`
   on checkout — including `\n` characters **inside** raw string literals.

4. The cloned working tree file now contains `\r\n` sequences within the raw
   string. The raw string literal at byte level is now larger than the logical
   Go string value.

5. When Go's parser processes this file:
   - `BasicLit.ValuePos` tracks the real file byte offset (includes `\r`)
   - `len(BasicLit.Value)` counts logical characters (excludes `\r`)
   - `BasicLit.End()` = `ValuePos + len(Value)` — **short by the `\r` count**
   - `genDecl.End()` points **inside** the raw string literal

## Mitigation

### User-side

Set `core.autocrlf=input` or `core.autocrlf=false`:
```bash
git config --global core.autocrlf false
```

### GitHub Actions

`actions/checkout@v4` already sets `core.autocrlf=input` for its own checkout
step, so CI builds are not affected.

### xgo-side

xgo should use `GenDeclSafeEnd()` (see `core_fix.diff` in the sibling issue)
instead of `genDecl.End()` to compute insertion positions, as a defense
against CRLF-corrupted files.
