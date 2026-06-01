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
