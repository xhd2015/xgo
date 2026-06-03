# GOCHAS.md — Common Gotchas in xgo Development

## Go Runtime

### GOTOOLCHAIN=local prevents unwanted toolchain downloads

Go 1.21+ auto-downloads a matching toolchain when `go.mod`'s `go` directive differs from the local binary. When rebuilding the compiler inside a patched GOROOT (where `src/cmd/go.mod` may specify a different version), set `GOTOOLCHAIN=local` to pin the build to the exact Go binary being executed.

### V1_0_0 guards legacy exec_tool interception

`V1_0_0 = false` (const, since v1.0.1.0) disables the old mechanism: building compiler at `instrumentDir/compile` and using `-toolexec` + exec_tool to intercept compiler calls. For go1.25+ file-based patches, the compiler is patched directly in GOROOT source — the go binary finds it natively via `GOROOT/pkg/tool/...`. Removing the guard is NOT the fix; the architecture changed.

## Go Compiler / Build

### Pre-compiled stdlib `.a` files are never rebuilt after source patching

Go uses pre-compiled `pkg/$GOOS_$GOARCH/*.a` files without checking source freshness (unless `-a` is passed). So patching the runtime source (adding `xgo_trap.go` with `XgoRegister`) has no effect — the stale `runtime.a` from the original GOROOT is still used. `functab.GetFuncs()` returns 0 entries because the test binary's runtime lacks `XgoRegister`.

**What `go install std` does:** compiles every standard library package from source and writes fresh `.a` files into `pkg/$GOOS_$GOARCH/`, replacing the originals. After this, `runtime.a` contains `XgoRegister` from the patched `xgo_trap.go`.

**Note:** The old programmatic path (`V1_0_0`) avoids this because it passes `-a` to `go test`, forcing recompilation from source at test time.

### Stale Go build cache survives source cleanup

After removing debug prints or other code from `patch/`, the old compiled objects may persist in Go's build cache (`$GOCACHE`, default `~/Library/Caches/go-build` on macOS). Even `go build -a` may not fully bypass it for all packages. The result: freshly built binaries still contain old (removed) code.

**Fix:** Clear the cache or use a fresh `GOCACHE` directory:
```sh
go clean -cache
# or
GOCACHE=/tmp/fresh-cache go build ...
```

## xgo Patch Syncing

### ApplyPatches order: `.xgo.patch` files must be applied BEFORE `generate` steps

`ApplyPatches` originally ran `filepath.Walk` (which applies `.xgo.patch` files) **after** the `generate` loop. Since `generate` includes `rebuild-compiler`, the compiler was built without the patched source (e.g., without `xgo_patch.Patch()` in `gc/main.go`). The compiler binary contained NO xgo rewrite code.

**Fix:** `instrument/patch/apply.go` — extracted `applyPatchFiles()` and called it before the generate loop.

**Symptom misdirection:** Manual shell `go build` appeared to produce a correct 27.9M binary while `exec.Command` produced a broken 27.8M one. This was a red herring — the manual test was run on an already-patched GOROOT from a prior run, masking the real bug.

### Run `go run ./script/generate` after modifying embedded sources

Several source directories are embedded into the xgo binary via `//go:embed`. Modifying them requires re-running generate before building xgo, otherwise the binary uses stale embedded copies.

| Source Dir | Asset Dir (embedded) | Command |
|---|---|---|
| `patches/` (versioned dirs like `go1.24/`, `go1.25/`) | `cmd/xgo/asset/patches/` | `go run ./script/generate cmd/xgo/asset/patches` |
| `patch/` (compiler link/syntax/instrument code) | `cmd/xgo/asset/compiler_patch_gen/` | `go run ./script/generate cmd/xgo/asset/compiler_patch_gen` |
| `runtime/` (runtime module code) | `cmd/xgo/asset/runtime_gen/` | `go run ./script/generate cmd/xgo/asset/runtime_gen` |

Sync both compiler-relevant assets at once:
```sh
go run ./script/generate cmd/xgo/asset/patches cmd/xgo/asset/compiler_patch_gen
```

Note: In development mode (`IS_DEV`), patches are loaded from the live source tree, so generate is only needed when building the release binary or running tests via `go run ./cmd/xgo` (which recompiles xgo).

## xgo Patch Generate Env

### Rebuild steps must set `GO_BYPASS_XGO=true` to prevent recursive interception

During GOROOT instrumentation, `apply.go` runs generate commands like `${INSTRUMENT_GOROOT}/bin/go build -a ... cmd/compile`. The go binary at `${INSTRUMENT_GOROOT}/bin/go` is the **instrumented** go binary — it has an `xgoPrecheck` hook that intercepts `build`/`test`/`run` and delegates to `xgo` in PATH.

Without `GO_BYPASS_XGO=true`, this creates a recursive loop: xgo runs `go build cmd/compile` → instrumented go delegates to xgo → xgo tries to re-instrument toolchain packages → variable trapping generates unresolved getter references (`undefined: src.NoXPos_xgo_get`).

**Fix:** Add `"env": {"GO_BYPASS_XGO": "true"}` to `rebuild-compiler`, `rebuild-stdlib`, and `rebuild-go` entries in `patches/go*/__config__.json`. Then run `go run ./script/generate cmd/xgo/asset/patches` to sync embedded assets.

The `xgoPrecheck` hook (`instrument/instrument_go/xgo_main_template.go:19`) checks `os.Getenv("GO_BYPASS_XGO") == "true"` and returns `false` (no interception) when set.

See also: `doc/development/WHY_DOUBLE_SETUP_EVER_FAILED.md`.

### `TestFuncNames` expects `init.func5` on go1.25 (not `init.func1`)

On go1.25, `needLocalRuntime=true` forces full xgo instrumentation (overlay + compiler rewrite + `compiler_extra.json`). The overlay injects 4 synthetic `func init()` blocks before user code, consuming `init.func1-4` and shifting user top-level closures by 4. The 4 stub template variables (`__xgo_register_1`, `__xgo_trap_1`, etc.) reference runtime dispatch functions (`runtime.XgoRegister`, `runtime.XgoTrapVar`, etc.) and are NOT themselves init closures. The test uses version-gated constants: `closure_name_go1.22_test.go` (go1.22–1.24: `init.func1`), `closure_name_go1.25_test.go` (go1.25+: `init.func5`).

See `doc/development/WHY_GO_1_25_FAIL_ON_TestFuncNames.md`.

## Go Version Differences

### Overlay stub indices (`__xgo_register_N`) shift with file count and declarations

xgo's overlay assigns a file index (the `N` in `__xgo_register_N`) based on zero-based position in `GoFiles + TestGoFiles + XTestGoFiles`. Every source file occupies a position, even build-tagged or declaration-free files. A package-level `const`/`var` in a `_test.go` file triggers var-trap instrumentation, injecting additional closures that shift the top-level closure numbering.

See `runtime/test/func_names/README.md` for a comparison table and validated points across 4 test packages.

### Go 1.25+ forbids overlay for files under GOMODCACHE

Go 1.25 introduced a restriction: the `-overlay` flag can no longer replace files located under `GOMODCACHE`. When an overlay contains such replacements, `go` exits with:

```
go: overlay contains a replacement for <file>. Files beneath GOMODCACHE (<path>) must not be replaced.
```

The check is in `src/cmd/go/internal/work/init.go:68-69`:
```go
if from, replaced := fsys.DirContainsReplacement(cfg.GOMODCACHE); replaced {
    base.Fatalf("go: overlay contains a replacement for %s. Files beneath GOMODCACHE (%s) must not be replaced.", from, cfg.GOMODCACHE)
}
```

`DirContainsReplacement` (`fsys/fsys.go:529`) checks if any overlay entry's path is within GOMODCACHE.

**Impact on xgo:** Since xgo runtime packages (like `runtime/trace`, `runtime/functab`) are Go module dependencies, they may reside in GOMODCACHE when resolved by `go test`. On go1.24, xgo modifies these files via overlay. On go1.25+, this is rejected.

**xgo's workaround** (`cmd/xgo/main.go:710`, commit `3123289`):
```go
needLocalRuntime := goVersion.Minor >= 25
```
When true, xgo:
1. Copies the xgo runtime to a local directory (`.xgo/gen/modules/`) outside GOMODCACHE
2. Modifies files directly on disk (not via overlay)
3. Uses a `replace` directive in a modified `go.mod` (`-modfile`) to resolve runtime from the local dir
4. Adds blank imports (`import _ "runtime/trace"`) to ensure runtime packages are compiled

**Side effect:** Step 4 forces `runtime/trace` into the test binary, where its instrumented code registers tracing functions in the functab. Tests that assert on exact functab contents need to account for this extra entry on go1.25+ (see `runtime/test/functab/` — uses `go_1_25.go`/`go_1_24.go` build-tag files with `IS_GO_25_OR_LATER` constant).

## macOS / Darwin

### Missing LC_UUID causes dyld abort on macOS 26+ with Go < 1.23

macOS 26+ (arm64) requires `LC_UUID` load command in executables. Go < 1.23 (and empirically go1.22.12) does not emit it, causing `dyld: missing LC_UUID load command ... signal: abort trap`. Fix: use `-ldflags=-linkmode=external` to invoke the system linker. See `instrument/build/build.go` (`NeedExternalLinker`/`ExternalLinkerFlags`) and `doc/development/LC_UUID_ERROR.md`.

## Windows

### `go build -o` does NOT append `.exe` on Windows — output file must specify it

When the `-o` flag is given to `go build` on Windows, Go writes the output to **exactly** the path specified in the argument. If the path is `.../pkg/tool/windows_amd64/compile` (no `.exe`), a file named `compile` is created WITHOUT the `.exe` extension. The existing `compile.exe` in the same directory is NOT overwritten.

The Go toolchain looks for `compile.exe` (with extension) when invoking the compiler during `go test` / `go build`. If `compile.exe` is the old unpatched binary — as it was merely copied from the original GOROOT by `syncGoroot` — the xgo link rewrite (`__xgo_trap` → `XgoTrap`) never runs, and `functab.GetFuncs()` returns 0 entries.

**Symptoms:**
- `functab.GetFuncs()` returns 0 on Windows, 600+ on Linux
- `TRAP_CHECK: interceptor did NOT fire`
- md5 of `compile.exe` in instrumented GOROOT matches the original
- File listing reveals two files: `compile` (correct/new) and `compile.exe` (old/original)

**Fix:** `instrument/patch/apply.go` — on Windows, after building the command args slice from the config, scan for `-o <path>` and append `.exe` if the path has no extension. Use `cmd_windows` (`CmdWindows` field) in `__config__.json` to provide a platform-specific args array with proper `.exe` extensions.

See `DEBUGGING.md` for the full investigation.

### Path separators in JSON config strings produce mixed `\` and `/` on Windows

`__config__.json` `cmd` fields use string format like `${INSTRUMENT_GOROOT}/bin/go`. The `${INSTRUMENT_GOROOT}` variable resolves to a path built by `filepath.Join` — which uses `\` on Windows. The literal `/bin/go` remains. The resulting path has mixed separators: `C:\Users\...\go1.25.10/bin/go`.

While Go's `os/exec` and filesystem handle this in many places, it creates fragile paths that can break in edge cases (e.g., `go build -o` may write to an unexpected location).

**Fix:** Use `cmd_windows` in `__config__.json` to specify a `[]string` args array with `\` separators and `.exe` extensions. Or ensure `substituteEnvSlice` performs `strings.ReplaceAll(arg, "/", string(filepath.Separator))` on each substituted arg when `runtime.GOOS == "windows"`.
