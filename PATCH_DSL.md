# Patch DSL

The **Patch DSL** is a declarative, AST-aware language for modifying Go source files. It is used to instrument the Go standard library for xgo's mocking, tracing, and interception features.

Patch files live under `patches/<go-version>/src/` and mirror the GOROOT directory structure. Files with the `.xgo.patch` extension are applied to their corresponding GOROOT source files at build time.

See [`patches/DESIGN_FILE_PATCH.md`](patches/DESIGN_FILE_PATCH.md) for internal architecture and design decisions.

## Block Structure

A `.xgo.patch` file contains one or more `<patch>` blocks:

```
<patch patch-name>
# comment (starting with #)
command
command

command
</patch>
```

- `#` starts a comment line (ignored).
- Blank lines are ignored.
- Each block has a unique name used for idempotent re-application.
- Blocks are applied sequentially; each block re-parses the file so later blocks see earlier modifications.

## Command Reference

### Positioning Commands

Move the cursor to a location in the source file. Each positioning command starts a new *sequence number* for stacking edits.

| Command | Description |
|---------|-------------|
| `goto struct <name>` | Cursor at the struct declaration |
| `goto func <name>` | Cursor at the start of a function declaration |
| `goto func (<receiver>) <name>` | Cursor at a method declaration |
| `goto interface <name>` | Cursor at an interface declaration |
| `goto opening {` | Cursor at the opening brace `{` of the current declaration (struct, func, interface) |
| `goto closing }` | Cursor at the closing brace `}` of the current declaration |
| `goto field <name>` | Cursor at a named field within the current struct (requires prior `goto struct`) |
| `match <text>` | Cursor at the start of matching text within current scope (for subsequent `insert_before`/`insert_after`) |
| `find_for_replace <text>` | Same as `match`, but selects a range for `replace` |

### Editing Commands

Modify the source at the cursor position.

| Command | Effect |
|---------|--------|
| `insert_before <text>` | Insert text at `cursor.offset` (before the cursor) |
| `insert_after <text>` | Insert text at `cursor.endOffset` (after the cursor) |
| `replace <text>` | Replace the range selected by a prior `find_for_replace` |
| `newline` | Append `\n` to the current accumulated text segment |
| `copy_func <source> as <target> append to file end` | Copy a function body, rename it, and append to the end of the file |

**Requirements:**
- `insert_before`, `insert_after`, `replace` all require non-empty text.
- `replace` requires a prior `find_for_replace` command.
- `goto opening/closing/field` requires a prior positioning command (struct, func, or interface).

## Examples

### Add a field to a struct

```
<patch xgo_add_field>
goto struct g
goto closing }
insert_before newField string
newline
</patch>
```

Before:
```go
type g struct {
    a int
}
```
After:
```go
type g struct {
    a int
    /*<begin xgo_add_field:2>*/newField string/*<end xgo_add_field:2>*/}
```

### Insert after a function's opening brace

```
<patch xgo_callback>
goto func goexit1
goto opening {
insert_after notifyCallback()
newline
</patch>
```

Before:
```go
func goexit1() {
    ...
}
```
After:
```go
func goexit1() {/*<begin xgo_callback:2>*/
    notifyCallback();/*<end xgo_callback:2>*/
    ...
}
```

### Insert before a matched line

```
<patch xgo_before_cmd>
goto func run
match cmd := exec.Command(name)
insert_before log.Printf("running: %v", name)
newline
</patch>
```

### Insert after a matched line

```
<patch xgo_mock_enable>
goto func runTest
match pkgArgs)
insert_after ;pkgs = xgoUnify(ctx, pkgs)
</patch>
```

### Replace a line

```
<patch xgo_linkname>
find_for_replace //go:linkname timeSleep time.Sleep
replace //go:linkname timeSleep time.runtimeSleep
</patch>
```

Before:
```go
//go:linkname timeSleep time.Sleep
func timeSleep(...)
```
After:
```go
/*<begin xgo_linkname:1><old://go:linkname timeSleep time.Sleep>*///go:linkname timeSleep time.runtimeSleep/*<end xgo_linkname:1>*/
func timeSleep(...)
```

### Replace a function signature

```
<patch xgo_sleep_wrap>
find_for_replace func Sleep(d Duration)
replace func XgoRealSleep(d Duration);func /*xgo_instr*/Sleep(d Duration){ XgoRealSleep(d); }
</patch>
```

### Copy a function to the end of the file

```
<patch add_real_now>
copy_func Now as XgoRealNow append to file end
</patch>
```

This copies the body of `Now()`, renames the function to `XgoRealNow()`, and appends it to the file end with markers.

### Insert a multi-line block (stacking `insert_after` + `newline`)

```
<patch xgo_noder_syntax>
goto func LoadPackage
match lines), "lines")
insert_after // auto gen
newline
insert_after if os.Getenv("FLAG")=="true" {
newline
insert_after     doWork()
newline
insert_after }
</patch>
```

Each `newline` appends `\n` to the preceding text segment.

## Stacking Behavior

When multiple editing commands operate at the **same cursor position** (same sequence number), they stack:

| Insert mode | Stack order | Example |
|-------------|-------------|---------|
| `insert_before` | **Reverse** (last command = closest to original code) | `insert_before A` then `insert_before B` → `AB` placed before the cursor |
| `insert_after` | **Forward** (first command = closest to original code) | `insert_after A` then `insert_after B` → `AB` placed after the cursor |
| `newline` | Appends `\n` to the preceding segment | `insert_after A` then `newline` → `A\n` |

Leading/trailing newlines from the stacking artifact are trimmed.

## Marker System & Idempotency

All edits are wrapped with **inline markers** so the system can:
- Detect existing edits and clear them before re-applying.
- Preserve original text for replace operations.

Markers **never** occupy their own line — they share lines with code to keep `panic file:line` references accurate.

### Insert Marker

```
/*<begin patch-name:seq>*/inserted-content/*<end patch-name:seq>*/
```

### Replace Marker

```
/*<begin patch-name:seq><old:original-text>*/replacement-content/*<end patch-name:seq>*/
```

### Clearing

- Before applying a patch, any existing markers with that patch's name are removed.
- For **insert** edits: markers and their content are deleted.
- For **replace** edits: the `/*<old:...>*/` content is restored.
- Applying the same patch twice has the same effect as applying it once.

## `__config__.json`

Each patch directory (e.g. `patches/go1.25/`) may have a `__config__.json` for directory copy and generate instructions:

```json
{
  "version": "go1.25+",
  "copy": [
    {"from": "relative/to/xgo_repo", "to": "relative/to/GOROOT"}
  ],
  "generate": [
    {
      "cmd": "shell command",
      "outputs": ["file/path"]
    }
  ]
}
```

- `copy`: copies a directory from the xgo repo into the instrumented GOROOT.
- `generate`: runs shell commands (supports `${VAR}` substitution from environment).
- Non-`.xgo.patch` files (except `__config__.json`) are copied one-to-one into GOROOT.

## Error Cases

| Condition | Error |
|-----------|-------|
| `insert_before` with no text | `insert_before requires text` |
| `insert_after` with no text | `insert_after requires text` |
| `replace` with no text | `replace requires text` |
| `replace` without `find_for_replace` | `replace requires prior find_for_replace` |
| Unknown goto target | `unknown goto target: "..."` |
| `match` not found in scope | `text not found in scope: "..."` |
| Struct/function/interface not found | `declaration not found: ...` / `function not found: ...` |
| Missing `</patch>` tag | `missing </patch> for "..."` |
| `copy_func` missing `as` keyword | `copy_func requires 'as' keyword` |

## Walkthrough: Adding a New Patch

1. **Identify the target** — find the GOROOT source file to modify and the exact location (struct, function, or text to match).

2. **Create the patch file** — mirror the GOROOT path under `patches/<go-version>/src/` with a `.xgo.patch` suffix.  
   Example: to patch `src/runtime/runtime2.go`, create `patches/go1.25/src/runtime/runtime2.go.xgo.patch`.

3. **Write the patch block** — use positioning commands to navigate, then editing commands to modify:
   ```
   <patch my_new_patch>
   # Describe what this patch does
   goto func targetFunc
   match theLine
   insert_after myExtraCode()
   </patch>
   ```

4. **Test** — run the xgo test suite to verify your changes:
   ```bash
   go test ./instrument/patch/... -v -count=1
   ```

5. **For new Go versions** — copy the previous version's patch directory and adjust selectors as needed:
   ```bash
   cp -r patches/go1.25 patches/go1.26
   ```

## Related

- [patches/DESIGN_FILE_PATCH.md](patches/DESIGN_FILE_PATCH.md) — architecture, design decisions, testing strategy
- [instrument/patch/apply_test.go](instrument/patch/apply_test.go) — comprehensive tests for parser and engine
