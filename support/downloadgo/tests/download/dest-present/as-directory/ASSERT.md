## Expected

- `Download` returns `$Dir/go1.19.13` with no library error.
- Stdout and Stderr are empty (no `download from`).
- Sentinel file is untouched.
- Hooks are unused (they panic if called).

## Side Effects

- Dest directory is not rewritten.

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
	assertEmptyWriters(t, resp)
	assertSentinelUntouched(t, req)
}
```
