# func_names tests

Test packages that verify xgo's overlay stub variable naming (`__xgo_register_N`, `__xgo_trap_N`, etc.) and closure numbering behavior across different file counts and declaration counts.

## Test packages

| Package | Files | Declarations | `__xgo_register` of func_names_test.go | Closure position | Validated point |
|---------|-------|-------------|----------------------------------------|------------------|-----------------|
| `pkg_with_only_test_files` | 1 .go (test only) | 4 stubs only | `_0` | none tested | Single test-only file: all stubs share index 0 |
| `pkg_with_source_files_with_4_files_with_4_decls` | 3 active + 1 build-tagged | 4 (const commented) | `_2` | func5 | Baseline: 2 stub-producing files, closure at func5 |
| `pkg_with_source_files_with_5_files_with_4_decls` | 4 active + 1 build-tagged | 4 (const commented) | `_3` | func5 | Adding a placeholder file shifts `__xgo_register` index but not closure position |
| `pkg_with_source_files_with_4_files_with_5_decls` | 3 active + 1 build-tagged | 5 (const active) | `_2` | func9 | A package-level `const` triggers var-trap instrumentation, injecting +4 closures |

## Key mechanics

- **File index** (`N` in `__xgo_register_N`): zero-based position in `GoFiles + TestGoFiles + XTestGoFiles`. Every file occupies a position even if excluded by build tags or lacking trap targets.
- **Closure numbering**: each `__xgo_register_N`, `__xgo_trap_N`, `__xgo_trap_var_N`, `__xgo_trap_varptr_N` anonymous function consumes one closure slot. Additional package-level `var`/`const` declarations trigger var-trap stubs that inject more closures.
- **Build-tagged files**: still occupy an index position regardless of whether they match the current Go version.
- **Stub generation**: files with no trap-worthy declarations (e.g., only `const` or blank `_` assignments) generate NO stubs, so their index position is "empty" — no `__xgo_register_N` variable exists for that index.
- **const caveat**: a package-level `const` in a `_test.go` file triggers the overlay's variable-trap instrumentation, injecting additional stub closures that shift the top-level closure numbering. To document the index without this side effect, the constant must be commented out (see `func_names_test.go` in each package).
