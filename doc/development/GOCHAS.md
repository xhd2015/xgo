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

## Go Version Differences

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
