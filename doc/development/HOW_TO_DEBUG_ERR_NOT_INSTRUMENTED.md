# How to Debug "Not Instrumented" Errors

When xgo tests fail with `functab.GetFuncs() returned 0 entries` or functions are not being trapped, the instrumentation pipeline is broken. Here's how to diagnose it.

## The Pipeline

```
xgo instruments user code → compiler Patch() runs link rewrite → runtime XgoRegister receives entries → functab populated
```

Each stage can fail independently.

## Step 1: Confirm the compiler patch runs

Add debug prints to `patch/patch.go` (see HOW_TO_DEBUG_COMPILER_ISSUE.md Step 3). Run the test and check if `Patch()` is called:

```sh
xgo test -v ./... 2>&1 | grep "XGO_PATCH"
```

If `Patch()` is never called, the instrumented compiler is not being used or is built without the patch. See GOCHAS.md § "ApplyPatches order".

## Step 2: Confirm link rewriting works

Check if `LinkXgoInit` finds `__xgo_init_*` functions and rewrites `__xgo_register_*` → `XgoRegister`:

```sh
xgo test -v ./... 2>&1 | grep "XGO_PATCH_LINK"
```

Expected output:
```
XGO_PATCH_LINK: LinkXgoInit: found __xgo_init_0
XGO_PATCH_LINK: linkStmts: __xgo_register_0 -> XgoRegister
XGO_PATCH_LINK: linkStmts: __xgo_trap_0 -> XgoTrap
```

If `LinkXgoInit` finds `__xgo_init_*` but `linkStmts` doesn't rewrite `__xgo_register_*`:
- `typecheck.LookupRuntime("XgoRegister")` returned nil — the runtime symbol doesn't exist
- The runtime hasn't been patched with `XgoRegister`

If no `__xgo_init_*` functions are found at all:
- xgo's instrumentation didn't generate registration code for the test package
- Check that the test package is listed in xgo's overlay/instrumentation scope

## Step 3: Check if the runtime has XgoRegister

```sh
INST=$(ls -d ~/.xgo/go-instrument/go1.25.10_*/go1.25.10)
grep "XgoRegister\|XgoSetupRegisterHandler" "$INST/src/runtime/xgo_trap.go"
```

Expected:
```
func XgoSetupRegisterHandler(register func(fn unsafe.Pointer)) {
func XgoRegister(fn interface{}) {
```

If `xgo_trap.go` is missing or these functions aren't there, the runtime patches weren't applied.

## Step 4: Check if runtime.a is rebuilt

Even with patched source, Go reuses pre-compiled `pkg/$GOOS_$GOARCH/runtime.a` from the original GOROOT. Verify:

```sh
# Check if stdlib was rebuilt (look for xgo symbols in the test binary)
go test -c -o /tmp/test.bin ./...
strings /tmp/test.bin | grep XgoRegister
```

If `XgoRegister` is not in the test binary, the runtime was not rebuilt from patched source. See GOCHAS.md § "Pre-compiled stdlib .a files".

## Step 5: Check functab init ordering

The `functab` package's `init()` registers a handler with `runtime.XgoSetupRegisterHandler()`:
```go
func init() {
    runtime.XgoSetupRegisterHandler(func(fn unsafe.Pointer) {
        RegisterFunc((*core.FuncInfo)(fn))
    })
}
```

This MUST run before the test package's `init()` (which calls `XgoRegister`). Go's init ordering follows import order — if `functab` is a transitive dependency of the test package, its init runs first.

Verify by printing in both `init()` functions:
```go
// In runtime/functab/functab.go
func init() {
    fmt.Fprintf(os.Stderr, "functab init called\n")
    ...
}
```

## Common Pitfalls

- **Pre-compiled runtime.a from original GOROOT**: See GOCHAS.md § "Pre-compiled stdlib .a files". Fix: `go install std` step in `__config__.json`.
- **Test binary compiled with original (unpatched) compiler**: If `-toolexec` is not set and `GOROOT` points to the original GOROOT, the instrumented compiler is never used. Check `GOROOT` and `PATH` in the test command's environment.
- **build cache masks the problem**: Use `-a` or `-count=1` to avoid cached test results.
