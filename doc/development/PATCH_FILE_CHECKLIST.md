
### Patched sources and affected files

When file-based patches are applied to a GOROOT, the following files are modified or created:

**Copy step** (from `patch/` → `src/cmd/compile/internal/xgo_rewrite_internal/patch/`):
| Source | Target (in instrumented GOROOT) |
|---|---|
| `patch/patch.go` | `src/cmd/compile/internal/xgo_rewrite_internal/patch/patch.go` |
| `patch/link/link_ir.go` | `src/cmd/compile/internal/xgo_rewrite_internal/patch/link/link_ir.go` |
| `patch/syntax/syntax.go` | `src/cmd/compile/internal/xgo_rewrite_internal/patch/syntax/syntax.go` |
| `patch/funcs/*.go` | `src/cmd/compile/internal/xgo_rewrite_internal/patch/funcs/` |
| `patch/info/*.go` | `src/cmd/compile/internal/xgo_rewrite_internal/patch/info/` |
| `patch/instrument/*.go` | `src/cmd/compile/internal/xgo_rewrite_internal/patch/instrument/` |
| `patch/ctxt/*.go` | `src/cmd/compile/internal/xgo_rewrite_internal/patch/ctxt/` |
| `patch/func_name/*.go` | `src/cmd/compile/internal/xgo_rewrite_internal/patch/func_name/` |

**`.xgo.patch` files** (applied to existing GOROOT source files):
| Patch | Affected GOROOT file |
|---|---|
| `src/cmd/compile/internal/gc/main.go.xgo.patch` | `src/cmd/compile/internal/gc/main.go` |
| `src/cmd/compile/internal/noder/noder.go.xgo.patch` | `src/cmd/compile/internal/noder/noder.go` |
| `src/cmd/cover/cover.go.xgo.patch` | `src/cmd/cover/cover.go` |
| `src/cmd/go/internal/load/test.go.xgo.patch` | `src/cmd/go/internal/load/test.go` |
| `src/cmd/go/internal/test/test.go.xgo.patch` | `src/cmd/go/internal/test/test.go` |
| `src/cmd/go/internal/work/exec.go.xgo.patch` | `src/cmd/go/internal/work/exec.go` |
| `src/cmd/go/main.go.xgo.patch` | `src/cmd/go/main.go` |
| `src/cmd/internal/test2json/test2json.go.xgo.patch` | `src/cmd/internal/test2json/test2json.go` |
| `src/encoding/json/encode.go.xgo.patch` | `src/encoding/json/encode.go` |
| `src/runtime/proc.go.xgo.patch` | `src/runtime/proc.go` |
| `src/runtime/runtime2.go.xgo.patch` | `src/runtime/runtime2.go` |
| `src/runtime/time.go.xgo.patch` | `src/runtime/time.go` |
| `src/testing/testing.go.xgo.patch` | `src/testing/testing.go` |
| `src/time/sleep.go.xgo.patch` | `src/time/sleep.go` |
| `src/time/time.go.xgo.patch` | `src/time/time.go` |

**Raw copied files** (new files added to GOROOT):
| Source | Target |
|---|---|
| `src/cmd/cover/xgo_cover.go` | `src/cmd/cover/xgo_cover.go` |
| `src/cmd/go/internal/test/xgo_testinfo.go` | `src/cmd/go/internal/test/xgo_testinfo.go` |
| `src/cmd/go/internal/test/xgo_testunified.go` | `src/cmd/go/internal/test/xgo_testunified.go` |
| `src/cmd/go/internal/work/xgo_work_sum.go` | `src/cmd/go/internal/work/xgo_work_sum.go` |
| `src/cmd/go/xgo_main.go` | `src/cmd/go/xgo_main.go` |
| `src/runtime/xgo_trap.go` | `src/runtime/xgo_trap.go` |

**Generate step outputs** (produced by commands in `__config__.json`):
| Kind | Produces |
|---|---|
| `mkbuiltin` | `src/cmd/compile/internal/typecheck/_builtin/runtime.go`, `src/cmd/compile/internal/typecheck/builtin.go` |
| `rebuild-compiler` | `pkg/tool/${GOOS}_${GOARCH}/compile` |
| `rebuild-stdlib` | all `pkg/${GOOS}_${GOARCH}/*.a` (standard library pre-compiled archives) |
| `rebuild-go` | `bin/go` |
