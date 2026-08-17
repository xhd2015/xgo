# Scenario

**Feature**: HTML with no version divs yields an empty list, not an error

```
FetchHTML(no id="go…" on a <div > line) -> List -> [] (no error)
```

## Preconditions

- Empty HTML and HTML without matching divs are the same outcome. This leaf
  uses non-empty HTML that looks version-like but does not match the parser:
  `<span id="go1.22.1">` (not a `<div ` line), `<div id="go">` (empty
  version), and `id="something-else"`.

## Steps

1. Inject `FetchHTML` returning that HTML.

```go
import (
	"context"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const noMatchHTML = `
<html><body>
<div>no version</div>
<div id="something-else">skip</div>
<span id="go1.22.1">not a div</span>
<div id="go"></div>
</body></html>
`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FetchHTML = func(ctx context.Context) (string, error) {
		return noMatchHTML, nil
	}
	return nil
}
```
