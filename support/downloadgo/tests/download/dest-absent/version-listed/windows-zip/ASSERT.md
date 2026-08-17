## Expected Output

```
---
version: 3
---
download from https://go.dev/dl/go1.19.13.windows-amd64.zip
```

Stdout is that one line, ending with a newline.

## Expected

- `Download` returns `$Dir/go1.19.13` with no library error.
- Marker `INSTALLED` is at that path.
- GetFile URL ends with `.windows-amd64.zip`.

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
	assertFetchRecorded(t, req, testVersionNaked, "windows", "amd64")
	assertStdoutDownloadFrom(t, resp.Stdout, archiveURL(testVersionNaked, "windows", "amd64"))
	if resp.Stderr != "" {
		t.Fatalf("stderr want empty, got %q", resp.Stderr)
	}
}
```
