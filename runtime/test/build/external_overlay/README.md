# External overlay composition — MECE test map

Feature under test: **xgo composes caller `-overlay` with generated instrument/prep overlays**
(A→B caller + B→C generated → A→C), with a single path identity in `overlayFS`.

Tests live as sibling packages under `runtime/test/build/` (discovered by `run-test`).
This file is the decision tree; package dirs are the leaves.

## DSN (participants / behaviors)

- **Caller** supplies `-overlay` (file redirects A→B).
- **xgo prep** may inject `import _ "…/runtime/trace"` into effective sources (`--strace`).
- **xgo instrument** rewrites effective sources (B→C content).
- **Go toolchain** builds with the final single-level overlay.
- **Test** asserts replacement **content** and that **mock/trap** works on overlaid funcs.

## MECE tree (significance order)

Split factor at each level is exclusive; siblings cover meaningful outcomes only.

```
L1: What must compose?                    (highest impact)
├── instrument_only/     → external_overlay_instrument
│     mock works on overlaid func; multi-version; clean sources
├── instrument + strace/ → external_overlay_strace*
│     blank import applied to effective (replacement) body
├── path_identity/       → external_overlay_path_identity
│     same file under path aliases still instruments
└── vet_import_boundary/ → external_overlay_composition  (go: 1.24)
      documents native Go overlay import resolution + xgo mock

L2 under instrument_only: overlay shape
├── single relative file     → external_overlay_instrument
└── multi-file overlay       → external_overlay_multifile

L2 under instrument + strace: which file prep vs overlay touch
├── same main package file   → external_overlay_strace
├── overlay sibling file     → external_overlay_strace_sibling
└── replacement already has trace import → external_overlay_strace_has_import

L2 under path_identity: how paths are spelled
├── absolute overlay JSON keys
└── project-dir via symlink (/var vs /private/var class on macOS)
```

## Leaf index

| Package | Proves | Notes |
|---------|--------|--------|
| `external_overlay_instrument` | A→B→C + mock | No Go 1.24 gate |
| `external_overlay_multifile` | Two mapped sources both instrumented | Relative keys |
| `external_overlay_strace` | Prep injects into overlaid main | `--strace` |
| `external_overlay_strace_sibling` | Prep on main; overlay on other file | Both survive |
| `external_overlay_strace_has_import` | Replacement already imports trace | No double-break |
| `external_overlay_path_identity` | Abs keys + symlink project-dir | Nested `xgo`/`go run ./cmd/xgo` |
| `external_overlay_composition` | Go 1.24+ vet + mock | Broken import on original |

Unit coverage (not under `runtime/test`): `instrument/overlay` for `absFileKey`, compose, blank-import-through-overlay.

## Non-goals (out of tree / covered elsewhere)

- Native Go-only vet matrix without xgo → `test/integrations/overlay_difference_…`
- Windows-only separator matrix → `instrument/overlay` unit tests (CI on windows)
- Overlay cycles → `instrument/overlay` unit tests
