# Design: File-Based, AST-Aware Patch System for xgo

*Design date: 2026-05-30*

## Overview

Starting with Go 1.25, xgo's patching system switches from programmatic text-anchor patching to file-based, AST-aware patch files stored in a version-specific directory structure. This makes patches more visible, reviewable, independent per Go version, and idempotent.

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
  "copy": {
    "src/cmd/compile/internal/xgo_rewrite_internal/patch/": "patch/"
  }
}
```

Key = GOROOT-relative destination. Value = xgo-repo-relative source directory.

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
| `replace <text>` | Cursor from `find_for_replace` | Replace cursor range with text |
| `newline` | - | Insert `\n` at current edit position |

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
/*<patch-name:seq>*/inserted-content/*<end>*/
```

**Replace marker** (stores original text for clearing):
```
/*<patch-name:seq><old:original-text>*/replacement-content/*<end>*/
```

Before applying a named patch, any existing markers with that name are cleared. For replace edits, clearing restores the `/*<old:...>*/` content. For insert edits, the markers and their content are removed.

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
→ `/*<xgo_g_field:1>*/__xgo_g __xgo_g;/*<end>*/}`

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
→ `/*<xgo_linkname:1><old://go:linkname timeSleep time.Sleep>*///go:linkname timeSleep time.runtimeSleep/*<end>*/`

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
   - Re-parse for next block

### `ApplyPatches(patchDir, goroot, xgoRepoRoot)`
Walks patch directory:
1. Load `__config__.json` for directory copy instructions
2. For each `.xgo.patch` file → apply to corresponding GOROOT file
3. For each other file → copy one-to-one to GOROOT with directories created

## Error Cases

| Condition | Error |
|---|---|
| `insert_before` with no text | `insert_before requires text` |
| `insert_after` with no text | `insert_after requires text` |
| `replace` with no text | `replace requires text` |
| `replace` without `find_for_replace` | `replace requires prior find_for_replace` |
| Unknown goto target | `unknown goto target: "..."` |
| `match` not found | `text not found in scope: "..."` |
| Declaration not found | `declaration not found: ...` |

## Testing Strategy

All tests in `instrument/patch/apply_test.go`:

### Parser Tests
- Parse valid `.xgo.patch` with multiple blocks
- Parse commands: goto, match, find_for_replace, insert_before, insert_after, replace, newline
- Handle comments and blank lines
- Error on unknown command

### Engine Tests
- `goto struct/func/interface` finds correct AST node
- `goto opening/closing` finds brace positions
- `goto field` finds named struct field
- `match` / `find_for_replace` finds text within scope
- `insert_before` inserts at correct offset
- `insert_before` multiple — reverse stacking
- `insert_after` multiple — forward stacking
- `replace` with old text preservation
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

Added to `cmd/xgo/main.go`:

```go
if useFilePatches && goVersion.Minor != 25 {
    return fmt.Errorf("--use-file-patches is only supported for go1.25, got go%d.%d", goVersion.Major, goVersion.Minor)
}
```

When enabled for go1.25+, `patchRuntime()` calls `ApplyPatches()` instead of the programmatic patching functions.

### New Go Version Workflow

When Go 1.26 arrives:
1. `cp -r patches/go1.25 patches/go1.26`
2. Build + test → find failures
3. Edit individual `.xgo.patch` files to adjust selectors
4. Run `go test ./instrument/patch/...` to verify

## Non-Goals

- Does NOT auto-generate patch files (maintainers copy and adjust)
- Does NOT migrate go1.17-1.24 patches (those remain programmatic)
- Does NOT handle merging or conflict resolution between patch versions
