# Scenario

**Feature**: matching `<div id="go…">` lines become naked versions in order

```
FetchHTML(three id="go…" divs) -> List -> ["1.22.1", "1.21.0", "1.20.3"]
```

## Preconditions

- HTML includes three version divs plus noise (`<div>` without id, and
  `id="something-else"`). Only `id="go…"` on a `<div ` line counts.

## Steps

1. Inject `FetchHTML` that returns the fixture HTML below.

```go
import (
	"context"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const matchingHTML = `
<html>
<body>
<div class="toggle" id="go1.22.1"></div>
<div id="go1.21.0"></div>
		<div id="go1.20.3"></div>
<div>no version</div>
<div id="something-else">skip</div>
</body>
</html>
`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FetchHTML = func(ctx context.Context) (string, error) {
		return matchingHTML, nil
	}
	return nil
}
```
