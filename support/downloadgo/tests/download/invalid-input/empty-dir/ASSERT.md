## Expected

- `Goroot` is empty.
- Stdout does not contain `download from`.

## Errors

- Library error is non-nil (empty `Dir` is rejected).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertHarnessOK(t, resp, err)
	if resp.Err == nil {
		t.Fatal("expected error for empty Dir")
	}
	if resp.Goroot != "" {
		t.Fatalf("Goroot = %q, want empty on empty Dir", resp.Goroot)
	}
	assertNoDownloadFrom(t, resp.Stdout)
}
```
