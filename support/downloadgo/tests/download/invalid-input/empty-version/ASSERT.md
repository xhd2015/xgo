## Expected

- `Goroot` is empty.
- Stdout does not contain `download from`.

## Errors

- Library error contains `download requires version`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertHarnessOK(t, resp, err)
	assertLibErrContains(t, resp, "download requires version")
	if resp.Goroot != "" {
		t.Fatalf("Goroot = %q, want empty on invalid version", resp.Goroot)
	}
	assertNoDownloadFrom(t, resp.Stdout)
}
```
