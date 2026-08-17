## Expected Output

```
---
version: 3
---
download from https://go.dev/dl/go1.19.13.linux-amd64.tar.gz
```

Stdout is that one line, ending with a newline.

## Expected

- `Download` returns `$Dir/go1.19.13` with no library error.
- Marker `INSTALLED` with contents `ok` is at that path.
- GetFile URL and dest match `go1.19.13.linux-amd64.tar.gz`.
- ListVersions, GetFile, and Extract each ran once.

## Side Effects

- Dummy archive written under `Dir`; extracted `go/` renamed to Target.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertHarnessOK(t, resp, err)
	assertNoLibErr(t, resp)
	want := wantGoroot(req.Dir, req.Version)
	if resp.Goroot != want {
		t.Fatalf("Goroot = %q, want %q", resp.Goroot, want)
	}
	assertMarkerAt(t, resp.Goroot)
	assertFetchRecorded(t, req, testVersionNaked, "linux", "amd64")
	assertStdoutDownloadFrom(t, resp.Stdout, archiveURL(testVersionNaked, "linux", "amd64"))
	if resp.Stderr != "" {
		t.Fatalf("stderr want empty, got %q", resp.Stderr)
	}
}
```
