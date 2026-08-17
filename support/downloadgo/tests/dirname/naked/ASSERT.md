## Expected

- `DirName("1.19.13")` returns exactly `go1.19.13`.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertHarnessOK(t, resp, err)
	assertNoLibErr(t, resp)
	if resp.Name != testDirName {
		t.Fatalf("DirName = %q, want %q", resp.Name, testDirName)
	}
}
```
