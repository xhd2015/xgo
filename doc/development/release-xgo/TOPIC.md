# Release xgo (tag + draft GitHub release)

Operational runbook for cutting a new xgo version: tag `vX.Y.Z`, optional `runtime/vX.Y.Z`, a **draft** GitHub release, and multi-arch binaries. For the short checklist only, see [CONTRIBUTING.md](../../../CONTRIBUTING.md#release-xgo).

## When to use

- Features/fixes are on (or about to land on) `master` and you need a versioned ship of the **xgo tool** (and sometimes the **runtime module**)
- You need tag `vX.Y.Z`, a **draft** GitHub release, and `xgo-release/` tarballs (plus `runtime/vX.Y.Z` when the runtime module is part of the ship)

## When not to use

- Local dev install only → `go run ./script/build-release --local` (see [INSTALLATION.md](../../INSTALLATION.md))
- Adding a new Go minor to xgo without shipping → use [patches/PROMPT_TEMPLATE.md](../../../patches/PROMPT_TEMPLATE.md) first
- Publishing without human review of release notes (stay on draft until a human clicks Publish)

## Prerequisites

- Work from **post-merge `master`** (or a short-lived `release-vX.Y.Z` branch that will push to `master`)
- Prefer a **dedicated worktree** so release build does not fight another checkout of `master`
- `gh` authenticated; rights to push tags and create releases
- Host Go able to build xgo (e.g. go1.25+)

## Version model

| Artifact | Role |
|----------|------|
| `cmd/xgo/version.go` | `VERSION`, `REVISION`, `NUMBER`, `CORE_VERSION`, `CORE_REVISION`, `CORE_NUMBER` |
| `runtime/core/version.go` | Runtime module version; kept in sync via generate / commit hooks when core fields change |
| `cmd/xgo/asset/runtime_gen/core/version.go` | Generated mirror of runtime version constants |
| Tag `vX.Y.Z` | Toolchain / root module tag (`go install …/cmd/xgo@v…`, GitHub release) |
| Tag `runtime/vX.Y.Z` | Go module tag for `github.com/xhd2015/xgo/runtime` — **only when that module ships** |
| `xgo-release/X.Y.Z/*.tar.gz` + `SHASUMS256.txt` | Build output (not committed) |
| GitHub **draft** release on `vX.Y.Z` | Human review before Publish |

### Release modes

Pick a mode before bumping versions. Dual-tag is **not** automatic for every ship.

| Situation | Bump `VERSION` | Bump `CORE_VERSION`? | Tag `vX.Y.Z` | Tag `runtime/vX.Y.Z` |
|-----------|----------------|----------------------|--------------|----------------------|
| **Tool-only** — `cmd/xgo`, patches, docs, or `runtime/test` only; published runtime lib unchanged; xgo does **not** need a newer runtime | Yes | **No** — leave previous | **Yes** | **No** |
| **Full (dual-tag)** — published runtime lib changed, and/or xgo depends on matching runtime / core identity users couple | Yes | **Yes** — set equal to `VERSION` | **Yes** | **Yes** |

**Published runtime** (counts toward full mode): packages users import under `runtime/` (`core`, `mock`, `trap`, …) and version constants when `CORE_*` is bumped.

**Does not force full mode by itself:** `runtime/test/…`, docs, CI, tooling-only paths. Those are still **tool-only** if the installable runtime module is unchanged.

Same conditional rule as [CONTRIBUTING.md — Release xgo](../../../CONTRIBUTING.md#release-xgo): tag `runtime/v…` **if there is a runtime update** (or a core sync that needs a new module version).

### `CORE_VERSION` rule

- Always bump `VERSION` for a release tag.
- Set `CORE_VERSION = VERSION` when this release changes **core** behavior that couples tool and runtime (runtime lib change, or xgo that requires a matching runtime / core identity — e.g. new Go minor support that needs both). That is the **full** mode case.
- Leave `CORE_VERSION` unchanged for **tool-only** ships (new flags, instrumentation UX, packaging that does not need a newer runtime). Leaving `CORE_VERSION` behind `VERSION` is intentional; the runtime compatibility check treats equal version strings as compatible.

When `CORE_VERSION == VERSION`, set `CORE_REVISION` / `CORE_NUMBER` to match `REVISION` / `NUMBER`, then run generate so runtime copies stay aligned.

## Procedure

Replace `X.Y.Z` with the next version (e.g. `1.2.3`). Choose **tool-only** or **full** first (see [Release modes](#release-modes)).

### 1. Land work on master

Merge feature PRs, then:

```sh
git fetch origin master
```

### 2. Release branch

If `master` is already checked out elsewhere:

```sh
git checkout -B release-vX.Y.Z origin/master
```

### 3. Bump versions

Edit `cmd/xgo/version.go`:

- Always set `VERSION = "X.Y.Z"`
- **Full mode:** also set `CORE_VERSION = "X.Y.Z"`
- **Tool-only:** leave `CORE_VERSION` / `CORE_REVISION` / `CORE_NUMBER` at the previous core values

### 4. Generate

```sh
go run ./script/generate
```

If `CORE_VERSION == VERSION` (full mode), ensure `CORE_REVISION` / `CORE_NUMBER` match `REVISION` / `NUMBER`. Fix manually and re-run generate if needed. Confirm:

```sh
rg 'const (VERSION|REVISION|NUMBER|CORE_)' cmd/xgo/version.go
rg 'const (VERSION|REVISION|NUMBER)' runtime/core/version.go
```

In tool-only mode, runtime core constants should still match the **previous** core version, not the new `VERSION`.

### 5. Release commit

```sh
git add -A
git commit -m "release vX.Y.Z"
```

Commit hooks update `REVISION` / `NUMBER` and, when core fields change, copy them into `runtime/core/version.go`.  
`REVISION` typically points at the **pre-release tip** with a `+1` suffix (same pattern as historical tags), not necessarily the release commit SHA itself.

### 6. Tag

Always create the toolchain tag:

```sh
git tag vX.Y.Z
```

**Full mode only** — also create the runtime module tag:

```sh
git tag runtime/vX.Y.Z
```

**Tool-only:** do **not** create `runtime/vX.Y.Z`.

### 7. Push commit and tags

```sh
git push origin HEAD:master   # or open a PR if branch protection requires it
git push origin vX.Y.Z
# full mode only:
git push origin runtime/vX.Y.Z
```

### 8. Draft GitHub release

**Full mode** notes (include runtime upgrade):

```sh
gh release create vX.Y.Z \
  --draft \
  --title "Xgo vX.Y.Z" \
  --notes-file - <<'EOF'
This release upgrades `xgo` from vPREV to **vX.Y.Z**

Major feature:
- …

To install `xgo` vX.Y.Z:

```sh
# upgrade xgo
go install github.com/xhd2015/xgo/cmd/xgo@vX.Y.Z

# update runtime
go get github.com/xhd2015/xgo/runtime@vX.Y.Z
```

For documentation, see https://github.com/xhd2015/xgo.

## What's Changed

* …

**Full Changelog**: https://github.com/xhd2015/xgo/compare/vPREV...vX.Y.Z
EOF
```

**Tool-only** notes (no runtime tag / no `go get` of a new runtime version):

```sh
gh release create vX.Y.Z \
  --draft \
  --title "Xgo vX.Y.Z" \
  --notes-file - <<'EOF'
This release upgrades `xgo` from vPREV to **vX.Y.Z**

Major feature:
- …

To install `xgo` vX.Y.Z:

```sh
go install github.com/xhd2015/xgo/cmd/xgo@vX.Y.Z
```

Runtime need not be upgraded for this release (existing `github.com/xhd2015/xgo/runtime` at the previous core version remains compatible).

For documentation, see https://github.com/xhd2015/xgo.

## What's Changed

* …

**Full Changelog**: https://github.com/xhd2015/xgo/compare/vPREV...vX.Y.Z
EOF
```

Keep the release **draft** until notes and binaries are reviewed.

### 9. Build binaries

```sh
go run ./script/build-release
```

Output layout:

```text
xgo-release/X.Y.Z/
  xgoX.Y.Z-darwin-amd64.tar.gz
  xgoX.Y.Z-darwin-arm64.tar.gz
  xgoX.Y.Z-linux-amd64.tar.gz
  xgoX.Y.Z-linux-arm64.tar.gz
  xgoX.Y.Z-linux-arm.tar.gz
  xgoX.Y.Z-windows-amd64.tar.gz
  xgoX.Y.Z-windows-arm64.tar.gz
  SHASUMS256.txt
```

### 10. Upload assets to the draft

```sh
gh release upload vX.Y.Z xgo-release/X.Y.Z/* --clobber
```

### 11. Human publish

Edit notes on GitHub if needed, then **Publish release**. Agents should stop at draft unless explicitly asked to publish.

## Verify

```sh
# constants
rg 'const (VERSION|CORE_VERSION)' cmd/xgo/version.go
rg 'const VERSION' runtime/core/version.go

# toolchain tag (always)
git show-ref vX.Y.Z
git log -1 --oneline vX.Y.Z

# full mode only — runtime module tag
git show-ref runtime/vX.Y.Z

# draft + assets
gh release view vX.Y.Z --json isDraft,url,assets
```

Expect `isDraft: true` until someone publishes; assets should list all tarballs plus `SHASUMS256.txt`.  
In tool-only mode, `runtime/vX.Y.Z` must **not** exist.

## Pitfalls

| Wrong | Correct |
|-------|---------|
| Tag without bumping `VERSION` | Bump `VERSION`, generate, then commit + tag |
| Always dual-tag even with no runtime / core change | **Tool-only:** `vX.Y.Z` only; dual-tag only in **full** mode |
| Skip `runtime/v…` after bumping `CORE_VERSION` and syncing `runtime/core/version.go` | **Full** mode: dual-tag so `go get …/runtime@v…` stays aligned |
| Tag `runtime/v…` without a real runtime/core ship | Prefer tool-only; do not invent a runtime module version |
| `CORE_VERSION == VERSION` but stale `CORE_REVISION` | Sync `CORE_REVISION` / `CORE_NUMBER` to `REVISION` / `NUMBER`, re-generate |
| Push tags only | Push the **release commit** to `master` (or via PR) **and** tags |
| Publish empty release then scramble for binaries | **Draft** → `build-release` → `gh release upload` → human publish |
| Tool-only notes that tell users to `go get runtime@vX.Y.Z` | Omit that line when there is no `runtime/vX.Y.Z` tag |
| Building in a dirty feature worktree | Use clean `origin/master` / `release-vX.Y.Z` worktree |

## Related

- [CONTRIBUTING.md — Release xgo](../../../CONTRIBUTING.md#release-xgo) — short checklist; same conditional runtime-tag rule
- [`script/build-release`](../../../script/build-release)
- [`script/generate`](../../../script/generate) (`cmd/xgo/version.go`, `runtime/core/version.go`)
- [patches/PROMPT_TEMPLATE.md](../../../patches/PROMPT_TEMPLATE.md) — add Go minor support **before** a release that advertises it
