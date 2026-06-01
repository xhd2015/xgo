# How to Debug Compiler Issues

## The Flow

xgo patches the Go compiler source in the instrumented GOROOT, then rebuilds:

```
patch files (.xgo.patch) → instrumented GOROOT source → go build → pkg/tool/<GOOS>_<GOARCH>/compile
```

The patched `gc/main.go` calls `xgo_patch.Patch()` at startup, which runs link rewriting on every compiled package.

## Step 1: Verify patches were applied to source

Check that `.xgo.patch` markers are present in the instrumented GOROOT:

```sh
INST=$(ls -d ~/.xgo/go-instrument/go1.25.10_*/go1.25.10)
grep "xgo_patch\|<begin call_xgo_patch>" "$INST/src/cmd/compile/internal/gc/main.go"
```

Expected output (line 222 is approximate):
```
5:package gc/*<begin import_xgo_patch>*/;import xgo_patch ...
222:ssagen.InitConfig()/*<begin call_xgo_patch>*/;xgo_patch.Patch()...
```

If markers are missing, `.xgo.patch` files weren't applied. See GOCHAS.md § "ApplyPatches order".

## Step 2: Check if the patch code directory exists

```sh
ls "$INST/src/cmd/compile/internal/xgo_rewrite_internal/patch/"
```

Should contain: `patch.go`, `link/`, `syntax/`, `funcs/`, etc.

## Step 3: Add debug prints to the compiler

Edit `patch/patch.go`:
```go
func Patch() {
    fmt.Fprintf(os.Stderr, "XGO_PATCH: Patch() called\n")
    linkFuncs()
}
```

Edit `patch/link/link_ir.go`:
```go
func LinkXgoInit(fn *ir.Func) {
    ...
    fnName := fn.Sym().Name
    if !strings.HasPrefix(fnName, "__xgo_init_") {
        return
    }
    fmt.Fprintf(os.Stderr, "XGO_PATCH: found %s\n", fnName)
    linkStmts(fn.Body, ...)
}
```

Then sync generate and rebuild:
```sh
go run ./script/generate cmd/xgo/asset/compiler_patch_gen
```

## Step 4: Manually rebuild the compiler

To isolate `exec.Command` vs shell behavior:

```sh
INST=$(ls -d ~/.xgo/go-instrument/go1.25.10_*/go1.25.10)
cd "$INST/src/cmd"
GOROOT="$INST" GOOS="" GOARCH="" GOTOOLCHAIN=local \
  "$INST/bin/go" build -a -o "$INST/pkg/tool/darwin_arm64/compile" cmd/compile
```

## Step 5: Verify the binary has patch code

```sh
strings "$INST/pkg/tool/darwin_arm64/compile" | grep -c "xgo_rewrite_internal/patch"
# Should be >0. Also check for debug strings:
strings "$INST/pkg/tool/darwin_arm64/compile" | grep "XGO_PATCH"
```

## Step 6: Compare binary sizes

| Size | Status |
|------|--------|
| ~27,832,754 | Broken — patch code not included |
| ~27,938,050 | Correct — patch code present |

## Step 7: Check if the compiler is actually invoked

Run a test with debug prints (from Step 3) and watch stderr:

```sh
xgo test --project-dir runtime/test -v -run TestFunctabMini ./functab 2>&1 | grep "XGO_PATCH"
```

If no output, the instrumented compiler is not being used. Check that `GOROOT` points to the instrumented GOROOT in the test command's environment.

## Common Pitfalls

- **`.xgo.patch` files applied after compiler rebuild**: ApplyPatches originally ran `filepath.Walk` after the `generate` loop. See GOCHAS.md § "ApplyPatches order".
- **Build cache hides the problem**: Use `-a` to force rebuild, or clear the Go build cache: `go clean -cache`.
- **Setup may skip rebuild**: If the instrumented GOROOT already exists, `syncGoroot` may do a partial sync and skip the compiler rebuild. Clear with `rm -rf ~/.xgo/go-instrument/go<VERSION>_*`.
