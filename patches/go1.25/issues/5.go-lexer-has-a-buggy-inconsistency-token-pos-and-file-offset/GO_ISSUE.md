# go/scanner: BasicLit.End() returns wrong position for raw string literals containing \r

**go version**: all (since Dec 2011)

**Package**: `go/scanner` / `go/ast`

## Summary

`go/scanner.scanRawString()` strips `\r` bytes from the returned literal value
(via `stripCR`, introduced in commit `fb6ffd8f787f`, CL 5495049) but does not
adjust `token.FileSet` byte offsets accordingly. This causes `ast.BasicLit.End()`
— computed as `ValuePos + len(Value)` — to return a position **short** by the
number of `\r` bytes inside the raw string.

In other words: `ValuePos` tracks real file byte offsets (including `\r`), but
`len(Value)` counts logical characters (excluding stripped `\r`). The computed
`End()` does not correspond to the actual end of the source token.

This affects any tool that relies on `ast.Node.End()` for source manipulation
(e.g., code generators, instrumentation tools). The discrepancy only manifests
on files where raw string literals contain carriage return bytes — which
happens naturally on Windows when git `core.autocrlf=true` injects `\r` into
all `\n` on checkout, including `\n` inside backtick-delimited raw strings.

## Root Cause

In `src/go/scanner/scanner.go`:

```go
// scanner.go:684-710
func (s *Scanner) scanRawString() string {
    offs := s.offset - 1
    hasCR := false
    for { ... if ch == '\r' { hasCR = true } ... }
    lit := s.src[offs:s.offset]     // raw bytes (include \r)
    if hasCR {
        lit = stripCR(lit, false)   // strip \r from value
    }
    return string(lit)
}

// scanner.go:667-682
func stripCR(b []byte, comment bool) []byte {
    c := make([]byte, len(b))
    i := 0
    for _, ch := range b {
        if ch != '\r' ... {
            c[i] = ch; i++
        }
    }
    return c[:i]
}
```

The returned *position* (`S.file.Pos(offs)`) counts raw bytes including `\r`.
The returned *literal value* has `\r` stripped. The caller in `go/ast` computes
`End()` from both halves — using incompatible byte counts:

```go
// go/ast/ast.go
func (x *BasicLit) End() token.Pos {
    return x.ValuePos + token.Pos(len(x.Value))
}
```

## Reproduction

The program below parses the same Go source with LF-only vs CRLF line endings.
Both parse to identical `Value` strings, but `BasicLit.End()` differs.

```go
package repro

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestBasicLitEndConsistency(t *testing.T) {
	srcLF := "package p\n\nconst help = `\na\nb\n`\n"
	srcCRLF := "package p\r\n\r\nconst help = `\r\na\r\nb\r\n`\r\n"

	tests := []struct {
		name            string
		src             string
		wantStartOffset int
		wantEndOffset   int // one past the closing backtick
	}{
		{"LF (\\n only)", srcLF, 24, 31},
		{"CRLF (\\r\\n)", srcCRLF, 26, 36},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tc.src, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}

			lit := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit)
			startOff := fset.Position(lit.ValuePos).Offset
			endOff := fset.Position(lit.End()).Offset

			// Find real closing backtick by scanning source bytes
			realEnd := -1
			for i := startOff + 1; i < len(tc.src); i++ {
				if tc.src[i] == '`' {
					realEnd = i + 1
					break
				}
			}

			gap := realEnd - endOff
			crCount := strings.Count(tc.src[startOff:realEnd], "\r")

			if startOff != tc.wantStartOffset {
				t.Errorf("startOffset: got %d, want %d", startOff, tc.wantStartOffset)
			}
			if endOff != tc.wantEndOffset {
				t.Errorf("endOffset: got %d, want %d", endOff, tc.wantEndOffset)
			}
			if gap != 0 {
				t.Errorf("gap: got %d, want 0 (crCount=%d)", gap, crCount)
			}
		})
	}
}
```

Save as `raw_string_test.go` and run with:

```bash
go test -v -run TestBasicLitEndConsistency
```

Expected output:

```
=== RUN   TestBasicLitEndConsistency
=== RUN   TestBasicLitEndConsistency/LF_(\n_only)
    raw_string_test.go:XX: LF (\n only): startOffset=24 endOffset=31 realEnd=31 gap=0 ✓
=== RUN   TestBasicLitEndConsistency/CRLF_(\r\n)
    raw_string_test.go:XX: CRLF (\r\n): startOffset=26 endOffset=33 realEnd=36 gap=3 ✗ FAIL: endOffset 33 != realEnd 36
--- FAIL: TestBasicLitEndConsistency
```

The LF and CRLF sources parse to the **same** `Value` string (`\r` stripped from
both), but the LF source has `End()` at the correct byte position while the CRLF
source has `End()` 3 bytes short — the 3 `\r` bytes that were stripped.

## Proposed Fix

One of:
1. Do not strip `\r` from the literal value (let callers decide)
2. Adjust the returned file position so `len(Value)` and `ValuePos` use the
   same byte count
3. Add a `RealEnd()` or similar API that scans source bytes to find the actual
   delimiter position

The test introduced with this change acknowledged the inconsistency:

```go
// scanner_test.go, from commit fb6ffd8f787f:
epos.Offset += len(e.lit) - len(lit) // correct position
```

## Additional Context

- Interpreted strings (`"..."`) are unaffected: `\r` alone is included in
  `len(Value)`, and `\r\n` causes a parse error (`newline in string`)
- Comments have their own `stripCR` call with consistent position tracking
- This has existed since Dec 15, 2011 (Go 1.0 era)
