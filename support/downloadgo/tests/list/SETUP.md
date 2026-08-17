# Scenario

**Feature**: `List` parses go.dev HTML through an injected `FetchHTML`

```
FetchHTML -> HTML
  -> List
  -> []string naked versions  |  error
```

## Preconditions

- `req.Op` is `list`.
- Production (`FetchHTML == nil`) would GET `https://go.dev/dl`. Every leaf
  here injects `FetchHTML` so the suite never hits the network.
- Parser (same as today's `parseDownloadVersions`): split on newlines, trim
  space, keep lines with prefix `<div `, then take the `id="go…"` value.

## Steps

1. Set `req.Op = "list"`.
2. Leaf injects `FetchHTML` (HTML or error).
3. `Run` calls `downloadgo.List`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "list"
	return nil
}
```
