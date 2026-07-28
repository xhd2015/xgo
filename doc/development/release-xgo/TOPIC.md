# Release xgo (tag + draft GitHub release)

Operational runbook for cutting a new xgo version: dual tags, draft GitHub release, multi-arch binaries. For the short checklist only, see [CONTRIBUTING.md](../../../CONTRIBUTING.md#release-xgo).

## When to use

- Features/fixes are on (or about to land on) `master` and you need a versioned ship
- You need tags `vX.Y.Z` + `runtime/vX.Y.Z`, a **draft** GitHub release, and `xgo-release/` tarballs

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
| `runtime/core/version.go` | Runtime module version; kept in sync via generate / commit hooks |
| `cmd/xgo/asset/runtime_gen/core/version.go` | Generated mirror of runtime version constants |
| Tags `vX.Y.Z` and `runtime/vX.Y.Z` | Toolchain tag + Go module tag (same version string) |
| `xgo-release/X.Y.Z/*.tar.gz` + `SHASUMS256.txt` | Build output (not committed) |
| GitHub **draft** release on `vX.Y.Z` | Human review before Publish |

**`CORE_VERSION` rule**

- Always bump `VERSION` for a release tag.
- Set `CORE_VERSION = VERSION` when the release changes **cmd/xgo and/or runtime** behavior users depend on (e.g. new Go minor support). That is the usual case.
- Leave `CORE_VERSION` unchanged only for rare non-core packaging bumps.

When `CORE_VERSION == VERSION`, set `CORE_REVISION` / `CORE_NUMBER` to match `REVISION` / `NUMBER`, then run generate so runtime copies stay aligned.

## Procedure

Replace `X.Y.Z` with the next version (e.g. `1.2.2`).

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

- Set `VERSION = "X.Y.Z"`
- Usually set `CORE_VERSION = "X.Y.Z"` as well

### 4. Generate

```sh
go run ./script/generate
```

If `CORE_VERSION == VERSION`, ensure `CORE_REVISION` / `CORE_NUMBER` match `REVISION` / `NUMBER`. Fix manually and re-run generate if needed. Confirm:

```sh
rg 'const (VERSION|REVISION|NUMBER|CORE_)' cmd/xgo/version.go
rg 'const (VERSION|REVISION|NUMBER)' runtime/core/version.go
```

### 5. Release commit

```sh
git add -A
git commit -m "release vX.Y.Z"
```

Commit hooks update `REVISION` / `NUMBER` and copy core fields into `runtime/core/version.go`.  
`REVISION` typically points at the **pre-release tip** with a `+1` suffix (same pattern as historical tags), not necessarily the release commit SHA itself.

### 6. Tag

For normal releases, create **both** tags:

```sh
git tag vX.Y.Z
git tag runtime/vX.Y.Z
```

### 7. Push commit and tags

```sh
git push origin HEAD:master   # or open a PR if branch protection requires it
git push origin vX.Y.Z runtime/vX.Y.Z
```

### 8. Draft GitHub release

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
rg 'const VERSION' cmd/xgo/version.go runtime/core/version.go

# tags on the release commit
git show-ref vX.Y.Z runtime/vX.Y.Z
git log -1 --oneline vX.Y.Z

# draft + assets
gh release view vX.Y.Z --json isDraft,url,assets
```

Expect `isDraft: true` until someone publishes; assets should list all tarballs plus `SHASUMS256.txt`.

## Pitfalls

| Wrong | Correct |
|-------|---------|
| Tag without bumping `VERSION` | Bump `VERSION` (and usually `CORE_VERSION`), generate, then commit + tag |
| `CORE_VERSION == VERSION` but stale `CORE_REVISION` | Sync `CORE_REVISION` / `CORE_NUMBER` to `REVISION` / `NUMBER`, re-generate |
| Push tags only | Push the **release commit** to `master` (or via PR) **and** tags |
| Publish empty release then scramble for binaries | **Draft** → `build-release` → `gh release upload` → human publish |
| Only one of `vX.Y.Z` / `runtime/vX.Y.Z` | Dual-tag for normal releases so `go install` and `go get …/runtime@v…` stay aligned |
| Building in a dirty feature worktree | Use clean `origin/master` / `release-vX.Y.Z` worktree |

## Related

- [CONTRIBUTING.md — Release xgo](../../../CONTRIBUTING.md#release-xgo)
- [`script/build-release`](../../../script/build-release)
- [`script/generate`](../../../script/generate) (`cmd/xgo/version.go`, `runtime/core/version.go`)
- [patches/PROMPT_TEMPLATE.md](../../../patches/PROMPT_TEMPLATE.md) — add Go minor support **before** a release that advertises it
