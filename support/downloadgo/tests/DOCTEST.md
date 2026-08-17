# downloadgo — prebuilt Go SDK downloader

Classic TDD (P1) doctests for greenfield package
`github.com/xhd2015/xgo/support/downloadgo`. **RED** until the implementer
extracts the library from `script/download-go`.

L2 in-process only: `Run` calls `DirName` / `Target` / `List` / `Download`.
No real go.dev, no curl/tar, no CLI (`script/download-go`), no `gotool/withgo`.

# DSN (Domain Specific Notion)

Importable helper that materializes a **prebuilt** Go SDK
(`go1.19.13.darwin-arm64.tar.gz` style) at `$Dir/goX.Y.Z`. Not a source
tarball, not `support/gitgoroot`.

**Participants**

- **Caller** — xgo tooling that wants a local GOROOT directory.
- **DirName / Target** — map a version string to `$Dir/go{naked}`.
- **List** — parse go.dev HTML via injected **FetchHTML**.
- **Download** — install that version into Target, writing progress to
  **Stdout** / **Stderr** on the options (never process stdio).
- **Seams** — **ListVersions**, **GetFile**, **Extract** on Download options;
  nil means production (go.dev list, curl, tar/zip). Tests always inject.

**Behaviors**

- Version `go1.19.13` or `1.19.13` → directory name always `go1.19.13`.
- **List**: trimmed lines starting with `<div ` that contain `id="go…"` yield
  naked versions in document order (`id="go1.22.1"` → `"1.22.1"`).
- **Download**, dest **directory** already there → return that path, no error,
  do not call ListVersions / GetFile / Extract, do not write `download from`.
- Dest exists as a **file** → error (not treated as installed).
- Empty version → error `download requires version`. Empty `Dir` → error.
- Dest missing → resolve GOOS/GOARCH (empty → runtime) → URL
  `https://go.dev/dl/go{ver}.{goos}-{goarch}{.tar.gz|.zip}` (`.zip` on
  windows) → list versions (list error: **warning on Stderr**, still proceed;
  list ok and version absent: `go%s not found`, no GetFile) → GetFile →
  Extract into a temp dir → rename extracted `go/` to Target. Stdout gets
  `download from <url>`.

## Locked API

Package `github.com/xhd2015/xgo/support/downloadgo`:

```text
func DirName(version string) string
func Target(dir, version string) string
func List(ctx context.Context, opts ListOptions) ([]string, error)
func Download(ctx context.Context, version string, opts Options) (goroot string, err error)

type Options struct {
    Dir, GOOS, GOARCH string
    Stdout, Stderr    io.Writer
    ListVersions      func(ctx context.Context) ([]string, error)
    GetFile           func(ctx context.Context, url, dest string) error
    Extract           func(archiveFile, destDir string) error
}

type ListOptions struct {
    FetchHTML func(ctx context.Context) (string, error)
}
```

Out of scope for this tree: `support/gitgoroot`, kool pin table, `Exec` /
`PATH` / `ModuleGoLine`, real network, rewriting curl/tar to `httputil`,
the `script/download-go` CLI.

## Version

0.0.2

## Decision Tree

```
support/downloadgo/tests/
├── dirname/                         # DirName: version spelling → dir name
│   ├── go-prefix/                   # "go1.19.13" → "go1.19.13"
│   └── naked/                       # "1.19.13"   → "go1.19.13"
├── target/                          # Target = Join(dir, DirName)
│   └── join-dir-and-version/
├── list/                            # List + injected FetchHTML
│   ├── matching-divs/               # three id="go…" → naked, in order
│   ├── no-matching-divs/            # empty / no version divs → []
│   └── fetch-error/                 # FetchHTML error → List error
└── download/                        # Download + injected seams
    ├── invalid-input/               # dest never consulted
    │   ├── empty-version/
    │   └── empty-dir/
    ├── dest-present/                # Target already on disk
    │   ├── as-directory/            # installed → return path; hooks unused
    │   └── as-file/                 # not installed → error
    └── dest-absent/                 # must fetch
        ├── version-listed/          # list includes version → GetFile+Extract
        │   ├── linux-targz/         # GOOS=linux  → .tar.gz URL
        │   └── windows-zip/         # GOOS=windows → .zip URL
        ├── version-absent/          # list omits version → go%s not found
        └── list-warn-proceed/       # list error → Stderr warning; still fetch
```

Parameter ranking (most → least significant):

1. **Operation** — `DirName` / `Target` / `List` / `Download`
2. **Download dest state** — invalid input / dest present / dest absent
3. **Present kind** — directory vs file; **absent list outcome** — listed /
   omitted / list error
4. **Archive suffix** — linux `.tar.gz` vs windows `.zip` (no GOOS matrix)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `dirname/go-prefix` | `DirName("go1.19.13")` → `go1.19.13` |
| 2 | `dirname/naked` | `DirName("1.19.13")` → `go1.19.13` |
| 3 | `target/join-dir-and-version` | `Target(dir, "go1.19.13")` → `dir/go1.19.13` |
| 4 | `list/matching-divs` | three `<div id="go…">` → naked versions in order |
| 5 | `list/no-matching-divs` | HTML with no version divs → empty slice, no error |
| 6 | `list/fetch-error` | `FetchHTML` error → `List` error, no versions |
| 7 | `download/invalid-input/empty-version` | empty version → `download requires version` |
| 8 | `download/invalid-input/empty-dir` | empty `Dir` → error |
| 9 | `download/dest-present/as-directory` | existing dir + sentinel; return path; hooks unused |
| 10 | `download/dest-present/as-file` | dest is a file → error; not installed |
| 11 | `download/dest-absent/version-listed/linux-targz` | listed; linux URL `.tar.gz`; marker at Target |
| 12 | `download/dest-absent/version-listed/windows-zip` | listed; windows URL `.zip`; marker at Target |
| 13 | `download/dest-absent/version-absent` | `go1.19.13 not found`; GetFile unused |
| 14 | `download/dest-absent/list-warn-proceed` | list error → Stderr warning; download still proceeds |

## How to Run

From the xgo module root (directory that contains `go.mod` for
`github.com/xhd2015/xgo`):

```sh
doctest vet ./support/downloadgo/tests
doctest test ./support/downloadgo/tests
```

Classic TDD: **`doctest vet` must pass**; **`doctest test` is expected RED**
until `support/downloadgo` exists with the locked API.

```go
import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/xgo/support/downloadgo"
)

// HookLog records injectable-seam calls for one leaf (per-request; parallel-safe).
type HookLog struct {
	ListCalls      int
	GetFileCalls   int
	ExtractCalls   int
	GetFileURL     string
	GetFileDest    string
	ExtractArchive string
	ExtractDest    string
}

// Request is one library call. Leaf Setup sets Op and the fields that Op reads.
type Request struct {
	Op      string // dirname | target | list | download
	Version string
	Dir     string
	GOOS    string
	GOARCH  string

	FetchHTML    func(ctx context.Context) (string, error)
	ListVersions func(ctx context.Context) ([]string, error)
	GetFile      func(ctx context.Context, url, dest string) error
	Extract      func(archiveFile, destDir string) error

	HookLog *HookLog

	SentinelPath string
	SentinelData string
}

// Response observes the library call. Library errors are in Err (not Run's error).
type Response struct {
	Name     string
	Target   string
	Goroot   string
	Versions []string
	Stdout   string
	Stderr   string
	Err      error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	ctx := context.Background()
	resp := &Response{}
	switch req.Op {
	case "dirname":
		resp.Name = downloadgo.DirName(req.Version)
	case "target":
		resp.Target = downloadgo.Target(req.Dir, req.Version)
	case "list":
		versions, err := downloadgo.List(ctx, downloadgo.ListOptions{
			FetchHTML: req.FetchHTML,
		})
		resp.Versions = versions
		resp.Err = err
	case "download":
		var stdout, stderr bytes.Buffer
		goroot, err := downloadgo.Download(ctx, req.Version, downloadgo.Options{
			Dir:          req.Dir,
			GOOS:         req.GOOS,
			GOARCH:       req.GOARCH,
			Stdout:       &stdout,
			Stderr:       &stderr,
			ListVersions: req.ListVersions,
			GetFile:      req.GetFile,
			Extract:      req.Extract,
		})
		resp.Goroot = goroot
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.Err = err
	default:
		return nil, fmt.Errorf("unknown op %q", req.Op)
	}
	return resp, nil
}
```
