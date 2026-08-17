## Expected

- `List` returns an empty slice (or nil).
- No library error.

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
	if len(resp.Versions) != 0 {
		t.Fatalf("versions = %#v, want empty", resp.Versions)
	}
}
```
