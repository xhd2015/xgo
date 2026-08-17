## Expected

- `List` returns `[]string{"1.22.1", "1.21.0", "1.20.3"}` in document order.
- Non-matching divs are ignored.

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
	assertVersions(t, resp.Versions, []string{"1.22.1", "1.21.0", "1.20.3"})
}
```
