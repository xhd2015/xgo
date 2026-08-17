## Expected

- `Versions` is empty.

## Errors

- Library error mentioning `html down`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertHarnessOK(t, resp, err)
	assertLibErrContains(t, resp, "html down")
	if len(resp.Versions) != 0 {
		t.Fatalf("versions = %#v, want empty on fetch error", resp.Versions)
	}
}
```
