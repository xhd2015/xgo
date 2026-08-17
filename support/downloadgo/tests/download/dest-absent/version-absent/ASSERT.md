## Expected

- `Goroot` is empty; Target was not created.
- Stdout does not contain `download from`.
- `ListVersions` ran once; GetFile was not called (it panics if it is).

## Errors

- Library error contains `go1.19.13 not found`.

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertHarnessOK(t, resp, err)
	assertLibErrContains(t, resp, "go1.19.13 not found")
	if resp.Goroot != "" {
		t.Fatalf("Goroot = %q, want empty when version is not listed", resp.Goroot)
	}
	assertNoDownloadFrom(t, resp.Stdout)
	if req.HookLog.ListCalls != 1 {
		t.Fatalf("ListVersions calls = %d, want 1", req.HookLog.ListCalls)
	}
	dest := wantGoroot(req.Dir, req.Version)
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("dest %s: %v (want not exist)", dest, statErr)
	}
}
```
