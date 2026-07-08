# Add Patches for New Go Version

Use this template when bumping xgo to the next stable Go release.

**Current status (as of go1.26 upgrade):** xgo supports `go1.17` ~ `go1.26`. File-based patches live under `patches/go1.24+`. The next target is **go1.27**.

Check whether the next version is available:

```sh
go run ./script/download-go list
```

Then add a new patch tree at `patches/<next-go-version>/` (e.g. `patches/go1.27/`).

Placeholders below:

- `<current-latest-go-version>` — latest supported minor today (e.g. `go1.26`)
- `<next-go-version>` — new minor to add (e.g. `go1.27`)

---

# Seed Patch and Prepare Source Code

## 1. Seed patches

```sh
mkdir ./patches/<next-go-version>
cp -R ./patches/
<current-latest-go-version>/__config__.json ./patches/<next-go-version>/__config__.json

cp -R ./patches/src<current-latest-go-version>/src ./patches/<next-go-version>/src
```

Update `patches/<next-go-version>/__config__.json`: bump `"version"` to `"<next-go-version>+"`.

## 2. Download upstream GOROOTs

xgo uses two GOROOT layouts:

| Directory | Purpose |
|-----------|---------|
| `go-release/` | `script/run-test` compatibility matrix |
| `go-release-git-versioned/` | integration tests, upstream diff comparison (git-backed) |

Download the new release into `go-release/`:

```sh
go run ./script/download-go <next-go-version>   # e.g. go1.27.0
```

Ensure the previous and new versions also exist under `go-release-git-versioned/`. Integration tests call `gitgoroot.EnsureGitGoroot` and will fetch missing versions on first use, or you can pre-fetch by running an integration test:

```sh
go run ./test/integrations/test_file_patch_can_be_repeated_on_patched_goroot --go-version <next-minor>
```

---

# Research and Adjust the Patches

Read `PATCH_DSL.md` to understand the patch DSL.

Compare upstream source between the two git-versioned trees, e.g.:

```sh
# example from go1.26 upgrade
diff -ru go-release-git-versioned/go1.25.10 go-release-git-versioned/go1.26.4 -- src/cmd/go src/cmd/compile src/runtime src/time src/testing src/encoding/json
```

Understand semantic changes in touched files, then check which `.xgo.patch` anchors or `.txt` copy files need adjustment.

**go1.26 example** (see `patches/go1.26/CHANGELOG`):

- `exec.go.xgo.patch` — anchor moved from `build` to `runCover`
- `test.go.xgo.patch` + `xgo_testunified.go` — `moduleLoaderState` plumbing changed

Present the planned patch changes and ask for user approval before implementing.

---

# Implement Beyond `patches/`

A new Go minor usually requires repo-wide version bumps. Checklist (go1.26 did all of these):

| Area | Files / action |
|------|----------------|
| Version constant | `support/goinfo/constants.go` — add `GO_VERSION_XX` |
| Supported range | `cmd/xgo/main.go` — `go1.17 ~ go1.XX` |
| Compiler adapter | `patch/legacy/adapter_go1.XX.go` (`//go:build go1.XX && !go1.XX+1`) |
| Prior adapter bounds | tighten `//go:build` on previous adapters (e.g. `go1.25 && !go1.27`) |
| Noder version map | `instrument/instrument_compiler/patch_noders.go` — extend upper bound |
| Runtime template guards | `instrument/instrument_runtime/template/{defs.go,func_name.go}` |
| Option tests | `cmd/xgo/option_test.go` if default/explicit behavior changes |
| Embedded assets | `go run ./script/generate cmd/xgo/asset/patches` |
| CI workflows | add `.github/workflows/go1-XX.yml` and `go1-XX-fast.yml` |
| Default CI Go | bump `go.yml`, `go-windows.yml` matrix to new version |
| Next preview | update `.github/workflows/go-next.yml` for `<next+1>` |

Fast workflow runs `go run ./script/run-test --log-debug -v --fast`.
Full workflow runs `go run ./script/run-test --install-xgo --with-setup --reset-instrument --log-debug -v`.

---

# Validate

## Unit / patch tests

```sh
go test ./instrument/patch/... -v -count=1
```

## Target version only

```sh
go run ./script/run-test --include <next-go-version>
```

`--include` uses prefix matching (`go1.26` matches `go1.26.4`).

If this fails, go back to patch research and retry.

## Integration tests (go1.24+)

These validate file-based patching against a git GOROOT:

```sh
go run ./test/integrations/test_file_patch_can_be_repeated_on_patched_goroot --go-version <next-minor>
go run ./test/integrations/test_file_patch_generated_same_diffs_as_programmatic_patch --go-version 1.24
```

**Linux CI pitfall (fixed in go1.26):** `test/integrations/internal/patch.go` must copy patch contents with `cp -R srcDir/. patchDir/`, not `cp -R srcDir/ patchDir/`. GNU `cp` nests `go1.XX/go1.XX/` when the destination is already named `go1.XX`, breaking patch rel paths.

**Subprocess tests:** `script/run-test` sets `XGO_TEST_COMMAND` via `buildXgoTestCommand()` so tests that spawn `xgo` use the in-tree binary, not a stale installed one.

---

# Regression Validation

Run integration baselines on prior file-patch versions to avoid regressions:

```sh
go run ./test/integrations/test_file_patch_can_be_repeated_on_patched_goroot --go-version 1.24
go run ./test/integrations/test_file_patch_can_be_repeated_on_patched_goroot --go-version 1.25
go run ./test/integrations/test_file_patch_can_be_repeated_on_patched_goroot --go-version 1.26
```

---

# CI Loop (on a PR branch)

After pushing, verify the new workflows on the PR:

```sh
sleep 10
github-fetch pr --logs 'https://github.com/xhd2015/xgo/pull/<PR>' --workflow 'Go 1-XX'
```

Note: `--workflow 'Go 1-25'` may match **Go 1-25 Fast** first. Use `gh run watch <run-id>` for the full workflow if needed.

If CI fails: fetch logs → reproduce locally → fix → `git add -A && git commit && git push` → poll again until green.

---

# Summarize to CHANGELOG

Document adjustments in `patches/<next-go-version>/CHANGELOG`:

- upstream versions compared
- which `.xgo.patch` / `.txt` files changed and why
- which patches were unchanged
- non-patch repo updates (adapters, constants, workflows, etc.)

See `patches/go1.26/CHANGELOG` for a completed example.