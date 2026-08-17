## Expected

- `Target("/tmp/installed", "go1.19.13")` equals
  `filepath.Join("/tmp/installed", "go1.19.13")`.

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertHarnessOK(t, resp, err)
	assertNoLibErr(t, resp)
	want := filepath.Join(req.Dir, testDirName)
	if resp.Target != want {
		t.Fatalf("Target = %q, want %q", resp.Target, want)
	}
}
```
