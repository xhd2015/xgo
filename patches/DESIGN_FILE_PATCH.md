# Design: File-Based, AST-Aware Patch System for xgo

*Design date: 2026-05-30*

## Overview

Starting Go 1.25, xgo's patching system switches from programmatic text-anchor patching to file-based, AST-aware patch files stored in a version-specific directory structure. This makes patches more visible, reviewable, independent per Go version, and idempotent.

The file-based patching is also back ported go 1.24.

## Directory Structure

```
patches/
  go1.25/
    __config__.json              # copy-dir instructions only
    src/
      runtime/
        xgo_trap.go              # new file, auto-copied → GOROOT/src/runtime/xgo_trap.go
        runtime2.go.xgo.patch   # applied  → GOROOT/src/runtime/runtime2.go
        proc.go.xgo.patch
      time/
        sleep.go.xgo.patch
      cmd/go/
        xgo_main.go              # new file, auto-copied
        main.go.xgo.patch
        internal/work/
          xgo_work_sum.go        # new file, auto-copied
          exec.go.xgo.patch
      cmd/cover/
        cover.go.xgo.patch
        xgo_cover.go
```

Rules:
- Directory structure **mirrors GOROOT** for direct file mapping
- `.xgo.patch` files → applied to corresponding file (strip suffix `.xgo.patch`)
- Other files (except `__config__.json`) → copied one-to-one
- `__config__.json` → only for directory copy instructions

## `__config__.json`

```jsonc
{
  "version": "go1.25+",
  "copy": [
    {"from": "patch/", "to": "src/cmd/compile/internal/xgo_rewrite_internal/patch/", "ignore_files": ["src/cmd/compile/internal/xgo_rewrite_internal/patch/go.mod"]}
  ],
  "generate": [
    {"kind": "rebuild-compiler", "cmd": "...", "outputs": []},
    {"kind": "rebuild-go", "cmd": "...", "outputs": []}
  ]
}
```

- `copy` is an **array of objects** with fields:
  - `from`: xgo-repo-relative source directory
  - `to`: GOROOT-relative destination
  - `ignore_files` (optional): file paths to remove after copy (relative to GOROOT)
- `generate` is an array of objects with fields:
  - `kind` (optional): category for selective skipping (e.g. `"rebuild-compiler"`)
  - `cmd`: shell command (supports `${VAR}` substitution)
  - `outputs`: list of output file paths (informational)

## `.xgo.patch` Format

### Block Structure

```
<patch patch-name>
# comment
command
command

command
</patch>
```

- `#` starts a comment line (ignored)
- Blank lines are allowed (ignored)
- Cursor resets at each `<patch>` boundary
- After each block, file is re-parsed (subsequent blocks see prior modifications)

### Command Reference

**Positioning commands** (each triggers a new sequence number):

| Command | Scope | Cursor Result |
|---|---|---|
| `goto struct <name>` | File-level | Point at struct declaration |
| `goto func <name>` | File-level | Point at function declaration |
| `goto func (<receiver>) <name>` | File-level | Point at method declaration |
| `goto interface <name>` | File-level | Point at interface declaration |
| `goto opening {` | Within current decl | Point at `{` (endOffset = pos+1) |
| `goto closing }` | Within current decl | Point at `}` (endOffset = pos+1) |
| `goto field <name>` | Within current struct | Point at named field |
| `match <text>` | Within current scope | Point at start of matched text (for insert) |
| `find_for_replace <text>` | Within current scope | Selects matched text range (for replace) |

**Editing commands** (all require non-empty text except `newline`):

| Command | Precondition | Effect |
|---|---|---|
| `insert_before <text>` | Cursor at a point | Insert text at `cursor.offset` |
| `insert_after <text>` | Cursor at a point | Insert text at `cursor.endOffset` |
| `insert_after_line <text>` | Cursor at a point | Like `insert_after`, but skips trailing `\n` so insert lands on next line |
| `replace <text>` | Cursor from `find_for_replace` | Replace cursor range with text |
| `replace_directive <old> with <new>` | Cursor from prior goto | Replace a compiler directive (e.g. `//go:linkname`) preserving its positional constraints |
| `newline` | - | Insert `\n` at current edit position |
| `copy_func <source> as <target> append to file end` | - | Copy a function body, rename it, and append to the end of the file |

### Consecutive Editing at Same Position

When multiple editing commands execute at the same cursor position (same sequence number), they stack:

- **`insert_before`**: stacks in **reverse** order (last command = closest to original code)
- **`insert_after`**: stacks in **forward** order (first command = closest to original code)
- **`newline`**: appends `\n` to the current accumulated text segment
- Leading/trailing newlines from the stacking artifact are trimmed

### Inline Markers (Idempotency)

All edits are wrapped with **inline, line-preserving** markers. Markers NEVER occupy their own line — they share lines with code. This ensures `panic file:line` references correspond to the original Go source.

**Insert marker**:
```
/*<begin patch-name>*/inserted-content/*<end patch-name>*/
```

**Replace marker** (stores original text for clearing):
```
/*<begin patch-name>*//*old:original-text*/replacement-content/*<end patch-name>*/
```

**Replace directive marker** (preserves line position of directives like `//go:linkname`):
```
/*<next-line-original patch-name>original-directive</next-line-original>*/
replacement-directive
```

Before applying a named patch, any existing markers with that name are cleared. For replace edits, clearing restores the `/*old:...*/` content. For `replace_directive` edits, clearing restores the `/*<next-line-original>*/` annotation's original text. For insert edits, the markers and their content are removed.

### Examples

**Add field to struct (`runtime2.go`)**:
```
<patch xgo_g_field>
goto struct g
goto closing }
insert_before __xgo_g __xgo_g;
newline
</patch>
```
→ `/*<begin xgo_g_field>*/__xgo_g __xgo_g;/*<end xgo_g_field>*/}`

**Insert at function start (`proc.go`)**:
```
<patch xgo_goexit1>
goto func goexit1
goto opening {
insert_after __xgo_callback_on_exit_g()
newline
</patch>
```

**Replace linkname (`time.go`)**:
```
<patch xgo_linkname>
find_for_replace //go:linkname timeSleep time.Sleep
replace //go:linkname timeSleep time.runtimeSleep
</patch>
```
→ `/*<begin xgo_linkname>*//*old://go:linkname timeSleep time.Sleep>*///go:linkname timeSleep time.runtimeSleep/*<end xgo_linkname>*/`

**Copy a function (`time.go`)**:
```
<patch xgo_real_now>
copy_func Now as XgoRealNow append to file end
</patch>
```

Copies the body of `Now()`, renames it to `XgoRealNow()`, and appends to the end of the file wrapped in markers.

**Replace a directive (`time.go`)**:
```
<patch xgo_runtime_linkname>
goto func timeSleep
replace_directive //go:linkname timeSleep time.Sleep with //go:linkname timeSleep time.XgoRealSleep
</patch>
```

Before:
```go
//go:linkname timeSleep time.Sleep
func timeSleep(...)
```
After:
```go
/*<next-line-original xgo_runtime_linkname>//go:linkname timeSleep time.Sleep</next-line-original>*/
//go:linkname timeSleep time.XgoRealSleep
func timeSleep(...)
```

## Engine Architecture

### `ParseXgoPatch(content) → *PatchFile`
Splits content on `<patch>`/`</patch>` tags, parses each block body into commands.

### `ApplyXgoPatchContent(source, patchContent) → string`
For each block:
1. `clearPatch(source, block.Name)` — remove any existing markers for this named patch
2. `applyPatch(source, block)` — execute commands against Go AST
   - Parse target with `go/parser.ParseFile`
   - Process commands: positioning → edit collection → marker wrapping
   - Apply edits in reverse offset order (preserves positions)
   - `replace_directive` uses a separate `/*<next-line-original>*/` annotation flow rather than standard begin/end markers
   - Re-parse for next block

### `ApplyPatches(patchDir, goroot, xgoRepoRoot, extraEnv, skipKinds)`
Walks patch directory:
1. Load `__config__.json` for directory copy and generate instructions
2. Process copy entries (directory copy with `ignore_files` support)
3. Process generate entries (run shell commands with `${VAR}` substitution, with `kind`-based skipping)
4. For each `.xgo.patch` file → apply to corresponding GOROOT file
5. For each other file → copy one-to-one to GOROOT with directories created

## Error Cases

| Condition | Error |
|---|---|
| `insert_before` with no text | `insert_before requires text` |
| `insert_after` with no text | `insert_after requires text` |
| `insert_after_line` with no text | `insert_after_line requires text` |
| `replace` with no text | `replace requires text` |
| `replace` without `find_for_replace` | `replace requires prior find_for_replace` |
| `replace_directive` with no old or new text | `replace_directive requires old and new text` |
| `replace_directive` missing `with` keyword | `replace_directive requires 'with' keyword: "..."` |
| Unknown goto target | `unknown goto target: "..."` |
| `match` not found | `text not found in scope: "..."` |
| Declaration not found | `declaration not found: ...` |
| Function not found | `function not found: ...` |
| `copy_func` missing `as` keyword | `copy_func requires 'as' keyword` |
| Unknown command | `unknown command: "..."` |

## Testing Strategy

All tests in `instrument/patch/apply_test.go`:

### Parser Tests
- Parse valid `.xgo.patch` with multiple blocks
- Parse commands: goto, match, find_for_replace, insert_before, insert_after, insert_after_line, replace, replace_directive, newline, copy_func
- Handle comments and blank lines
- Error on unknown command

### Engine Tests
- `goto struct/func/interface` finds correct AST node
- `goto opening/closing` finds brace positions
- `goto field` finds named struct field
- `match` / `find_for_replace` finds text within scope
- `insert_before` inserts at correct offset
- `insert_before` multiple — reverse stacking
- `insert_after` / `insert_after_line` — forward stacking; `insert_after_line` skips trailing newline
- `replace` with old text preservation
- `replace_directive` produces `/*<next-line-original>*/` annotation
- `copy_func` copies and renames function body
- Multiple `<patch>` blocks — second sees first's changes

### Marker Tests
- Markers are inline (no extra marker-only lines)
- `clearPatch` removes insert markers
- `clearPatch` restores replace old content
- Idempotent: apply twice = apply once
- Idempotent with changes: apply v2 after v1 = v2 only

### Error Tests
- `insert_before` empty text → error
- `replace` without `find_for_replace` → error
- `match` not found → error
- `goto` unknown target → error

```bash
go test ./instrument/patch/... -v -count=1
```

## Integration

### `--use-file-patches` flag

Added to `cmd/xgo/main.go`. When enabled for go1.24+, `patchRuntime()` calls `ApplyPatches()` instead of the programmatic patching functions.

### New Go Version Workflow

When Go 1.26 arrives:
1. `cp -r patches/go1.25 patches/go1.26`
2. Build + test → find failures
3. Edit individual `.xgo.patch` files to adjust selectors
4. Run `go test ./instrument/patch/...` to verify

## Non-Goals

- Does NOT auto-generate patch files (maintainers copy and adjust)
- Does NOT migrate go1.17-1.23 patches (those remain programmatic)
- Does NOT handle merging or conflict resolution between patch versions
- File-based patches currently support go1.24+
