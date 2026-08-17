## Expected

- `Goroot` is empty (not treated as installed).
- Stdout does not contain `download from`.
- The dest path is still a regular file.

## Errors

- Library error is non-nil.

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertHarnessOK(t, resp, err)
	if resp.Err == nil {
		t.Fatal("expected error when dest exists as a file")
	}
	if resp.Goroot != "" {
		t.Fatalf("Goroot = %q, dest-as-file must not be treated as installed", resp.Goroot)
	}
	assertNoDownloadFrom(t, resp.Stdout)
	dest := wantGoroot(req.Dir, req.Version)
	fi, statErr := os.Stat(dest)
	if statErr != nil {
		t.Fatalf("stat dest: %v", statErr)
	}
	if fi.IsDir() {
		t.Fatal("dest was replaced by a directory")
	}
}
```
