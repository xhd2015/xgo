## Expected Output

Stdout:

```
---
version: 3
---
download from https://go.dev/dl/go1.19.13.linux-amd64.tar.gz
```

Stderr (existing CLI warning text, extracted):

```
---
version: 3
---
WARNING cannot get go version list:list down
```

Each stream is one line ending with a newline.

## Expected

- `Download` returns `$Dir/go1.19.13` with no library error.
- Marker is present; GetFile + Extract ran once.

## Errors

- None from `Download` (list error is only a warning).

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
	assertStderrListWarning(t, resp.Stderr, "list down")
}
```
