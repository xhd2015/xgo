# Scenario

**Feature**: in-process prebuilt Go SDK downloader (`support/downloadgo`)

```
Caller
  -> DirName / Target                 -> go{naked} path
  -> List(FetchHTML)                  -> naked versions
  -> Download(Dir, version, hooks)    -> $Dir/goX.Y.Z
# tests inject FetchHTML / ListVersions / GetFile / Extract — never go.dev
```

## Preconditions

- Package `github.com/xhd2015/xgo/support/downloadgo` exports the locked API
  in root `DOCTEST.md`.
- Leaves run under `t.Parallel()` in one process. **No** `t.Setenv`,
  `t.Chdir`, `os.Setenv`, `os.Chdir`, or rewriting `os.Stdout` / `os.Stderr`.
- Dest directories come from `t.TempDir()` (writable, isolated). Leaf source
  under `d.DOCTEST_CASE` is not used as an install dir.
- Every `Download` leaf injects `ListVersions` / `GetFile` / `Extract` (panic
  or record). Production nil hooks are never exercised here.
- `List` leaves inject `FetchHTML`. No live `https://go.dev/dl`.

## Steps

1. Root `Setup` allocates a per-request `HookLog`.
2. Grouping `Setup` sets `req.Op` and shared dest / hook policy.
3. Leaf `Setup` fills the one scenario's version, dest, HTML, or GOOS.
4. Root `Run` dispatches to `DirName` / `Target` / `List` / `Download`.
5. Leaf `Assert` checks the returned path, versions, writers, and hook log.

## Context

- Archive URL: `https://go.dev/dl/go{naked}.{goos}-{goarch}{.tar.gz|.zip}`.
- Windows suffix is `.zip`; every other GOOS is `.tar.gz`.
- Extracted archive layout is `destDir/go/…`; Download renames `go/` to Target.
- Success marker written by test `Extract`: `INSTALLED` with contents `ok`.
- Dest-present sentinel: `SENTINEL` with contents `keep-me`.
- List warning (existing CLI text, extracted):
  `WARNING cannot get go version list:%v` on Stderr.

```go
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

const (
	testVersionNaked  = "1.19.13"
	testVersionGo     = "go1.19.13"
	testDirName       = "go1.19.13"
	installedMarker   = "INSTALLED"
	installedContents = "ok"
	sentinelName      = "SENTINEL"
	sentinelContents  = "keep-me"
	listWarnPrefix    = "WARNING cannot get go version list:"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.HookLog == nil {
		req.HookLog = &HookLog{}
	}
	return nil
}

func goDirName(version string) string {
	if strings.HasPrefix(version, "go") {
		return version
	}
	return "go" + version
}

func wantGoroot(dir, version string) string {
	return filepath.Join(dir, goDirName(version))
}

func archiveURL(naked, goos, goarch string) string {
	suffix := ".tar.gz"
	if goos == "windows" {
		suffix = ".zip"
	}
	return "https://go.dev/dl/go" + naked + "." + goos + "-" + goarch + suffix
}

func archiveBase(naked, goos, goarch string) string {
	suffix := ".tar.gz"
	if goos == "windows" {
		suffix = ".zip"
	}
	return "go" + naked + "." + goos + "-" + goarch + suffix
}

func writeInstalledSDK(destDir string) error {
	goDir := filepath.Join(destDir, "go")
	if err := os.MkdirAll(goDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(goDir, installedMarker), []byte(installedContents), 0644)
}

func installPanicHooks(req *Request) {
	req.ListVersions = func(ctx context.Context) ([]string, error) {
		panic("ListVersions must not be called")
	}
	req.GetFile = func(ctx context.Context, url, dest string) error {
		panic("GetFile must not be called")
	}
	req.Extract = func(archiveFile, destDir string) error {
		panic("Extract must not be called")
	}
}

func installListVersions(req *Request, versions []string, listErr error) {
	req.ListVersions = func(ctx context.Context) ([]string, error) {
		req.HookLog.ListCalls++
		if listErr != nil {
			return nil, listErr
		}
		out := make([]string, len(versions))
		copy(out, versions)
		return out, nil
	}
}

func installSuccessIO(req *Request) {
	req.GetFile = func(ctx context.Context, url, dest string) error {
		req.HookLog.GetFileCalls++
		req.HookLog.GetFileURL = url
		req.HookLog.GetFileDest = dest
		return os.WriteFile(dest, []byte("dummy-archive"), 0644)
	}
	req.Extract = func(archiveFile, destDir string) error {
		req.HookLog.ExtractCalls++
		req.HookLog.ExtractArchive = archiveFile
		req.HookLog.ExtractDest = destDir
		return writeInstalledSDK(destDir)
	}
}

func installPanicGetFileExtract(req *Request) {
	req.GetFile = func(ctx context.Context, url, dest string) error {
		panic("GetFile must not be called")
	}
	req.Extract = func(archiveFile, destDir string) error {
		panic("Extract must not be called")
	}
}

func assertHarnessOK(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("harness Run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
}

func assertNoLibErr(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("unexpected library error: %v", resp.Err)
	}
}

func assertLibErrContains(t *testing.T, resp *Response, substr string) {
	t.Helper()
	if resp.Err == nil {
		t.Fatalf("expected library error containing %q, got nil", substr)
	}
	if !strings.Contains(resp.Err.Error(), substr) {
		t.Fatalf("library error %q, want substring %q", resp.Err.Error(), substr)
	}
}

func assertVersions(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("versions = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("versions[%d] = %q, want %q (full %#v)", i, got[i], want[i], got)
		}
	}
}

func assertMarkerAt(t *testing.T, goroot string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(goroot, installedMarker))
	if err != nil {
		t.Fatalf("read installed marker: %v", err)
	}
	if string(data) != installedContents {
		t.Fatalf("marker contents %q, want %q", string(data), installedContents)
	}
}

func assertSentinelUntouched(t *testing.T, req *Request) {
	t.Helper()
	data, err := os.ReadFile(req.SentinelPath)
	if err != nil {
		t.Fatalf("read sentinel: %v", err)
	}
	if string(data) != req.SentinelData {
		t.Fatalf("sentinel %q, want %q (dest was overwritten)", string(data), req.SentinelData)
	}
}

func assertNoDownloadFrom(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "download from") {
		t.Fatalf("stdout must not contain download from:\n%s", stdout)
	}
}

func assertStdoutDownloadFrom(t *testing.T, stdout, url string) {
	t.Helper()
	line := regexp.QuoteMeta("download from " + url)
	assert.Output(t, stdout, fmt.Sprintf("---\nversion: 3\n---\n%s\n", line))
}

func assertStderrListWarning(t *testing.T, stderr, errText string) {
	t.Helper()
	line := regexp.QuoteMeta(listWarnPrefix + errText)
	assert.Output(t, stderr, fmt.Sprintf("---\nversion: 3\n---\n%s\n", line))
}

func assertEmptyWriters(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Stdout != "" {
		t.Fatalf("stdout want empty, got %q", resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr want empty, got %q", resp.Stderr)
	}
}

func assertFetchRecorded(t *testing.T, req *Request, naked, goos, goarch string) {
	t.Helper()
	wantURL := archiveURL(naked, goos, goarch)
	wantDest := filepath.Join(req.Dir, archiveBase(naked, goos, goarch))
	if req.HookLog.ListCalls != 1 {
		t.Fatalf("ListVersions calls = %d, want 1", req.HookLog.ListCalls)
	}
	if req.HookLog.GetFileCalls != 1 {
		t.Fatalf("GetFile calls = %d, want 1", req.HookLog.GetFileCalls)
	}
	if req.HookLog.ExtractCalls != 1 {
		t.Fatalf("Extract calls = %d, want 1", req.HookLog.ExtractCalls)
	}
	if req.HookLog.GetFileURL != wantURL {
		t.Fatalf("GetFile url = %q, want %q", req.HookLog.GetFileURL, wantURL)
	}
	if req.HookLog.GetFileDest != wantDest {
		t.Fatalf("GetFile dest = %q, want %q", req.HookLog.GetFileDest, wantDest)
	}
	if req.HookLog.ExtractArchive != wantDest {
		t.Fatalf("Extract archive = %q, want %q", req.HookLog.ExtractArchive, wantDest)
	}
}
```
