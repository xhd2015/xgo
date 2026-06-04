---
os: all
version: all
upstream: https://go-review.googlesource.com/c/go/+/5495049
---

# go/scanner: Buggy Inconsistency Between token.Pos and File Offset in Raw String Literals

## Summary

`go/scanner.scanRawString()` strips `\r` (carriage return, 0x0D) bytes from the
**literal value** but does NOT adjust the **file position** accordingly. This
causes `go/ast.BasicLit.End()` (computed as `ValuePos + len(Value)`) to return
a position that is **short** by the number of `\r` bytes inside the raw string.

This inconsistency has existed since **December 15, 2011** (commit
`fb6ffd8f787f` in the Go repository).

## Root Cause

In `src/go/scanner/scanner.go`, the `scanRawString()` function:

```go
// go/scanner/scanner.go:684-710
func (s *Scanner) scanRawString() string {
    offs := s.offset - 1

    hasCR := false
    for {
        ch := s.ch
        if ch < 0 { break }
        s.next()
        if ch == '`' { break }
        if ch == '\r' {
            hasCR = true                     // (1) track \r presence
        }
    }

    lit := s.src[offs:s.offset]              // (2) raw bytes include \r
    if hasCR {
        lit = stripCR(lit, false)            // (3) strip \r from VALUE
    }
    return string(lit)                       // (4) return stripped value
}
```

The *position* returned to the caller is based on `s.file.Pos(offs)` which
counts raw bytes **including** `\r`. But the *literal value* has `\r` stripped.

This creates a split:
- `BasicLit.ValuePos` — tracks real byte offsets (with `\r`)
- `len(BasicLit.Value)` — counts characters post-strip (without `\r`)
- `BasicLit.End() = ValuePos + len(Value)` — **wrong** by `\r` count

## `stripCR` Function

```go
// go/scanner/scanner.go:667-682
func stripCR(b []byte, comment bool) []byte {
    c := make([]byte, len(b))
    i := 0
    for j, ch := range b {
        if ch != '\r' || ... {
            c[i] = ch
            i++
        }
    }
    return c[:i]
}
```

## Trigger

This bug is only triggered when raw string literals (backtick-delimited)
contain `\r` bytes. This happens naturally on Windows when:

1. Git for Windows defaults to `core.autocrlf=true`
2. On `git clone`/`git checkout`, all `\n` → `\r\n`
3. `\r` bytes land **inside** backtick-delimited raw strings that span
   multiple lines (e.g., help text constants)

Interpreted string literals (`"..."`) are unaffected because:
- `\r` alone inside `"..."` is valid and included in `len(Value)` → no gap
- `\r\n` inside `"..."` is a syntax error (`newline in string`) → file won't parse

## Reproduction

```bash
go run ./patches/go1.25/issues/go-lexer-has-a-buggy-inconsistency-token-pos-and-file-offset/repro
```

Output:
```
=== LF (\n only) ===
  Value=`\nUsage:\n  -func match\n`  len(Value)=24
  End()=offset correctly after closing `
  ✓ End() CORRECT (gap=0)

=== CRLF (\r\n) ===
  Value=`\nUsage:\n  -func match\n`  len(Value)=24  (same!)
  End()=offset short by \r count
  ✗ End() WRONG by 3 bytes
```

## Impact on xgo

xgo instruments Go source files by inserting stub declarations at positions
computed from `GenDecl.End()`. When the file has `\r` bytes inside raw strings:

- `genDecl.End()` → `BasicLit.End()` → points INSIDE the raw string
- Inserted stubs land **inside** the raw string literal, invisible to the compiler
- Results in `undefined:` and `syntax error` compilation failures

## Git History

```bash
cd <go-repo>
git log --oneline --all -S 'stripCR' -- src/go/scanner/scanner.go
# fb6ffd8f78 go/scanner: strip CRs from raw literals
# 7b9a6d8dda go/scanner: strip carriage returns from comments
```

Commit `fb6ffd8f787f` (Dec 15, 2011, Robert Griesemer) is the introduction point.
The test file adjustment reveals they were aware of the position discrepancy but
only corrected it in tests, not in the API:

```go
// scanner_test.go, from the commit:
epos.Offset += len(e.lit) - len(lit) // correct position
```

## Workaround

When parsing source files, already replace "\r\n" with "\n"

## Scope

| Literal type | `\r` inside? | `End()` wrong? |
|-------------|-------------|---------------|
| Raw string `` `...` `` | Yes (CRLF injection) | **Yes** |
| Interpreted `"..."` | `\r` only: valid | No |
| Interpreted `"..."` | `\r\n`: syntax error | Can't parse |
| Comments `//`, `/* */` | Stripped by separate code | No |

## Related

- [windows-raw-string-syntax-rewrite-issue](../windows-raw-string-syntax-rewrite-issue/) — xgo bug caused by this inconsistency
- [windows-git-checkout-introduces-cr](../windows-git-checkout-introduces-cr/) — why `\r` appears in files
