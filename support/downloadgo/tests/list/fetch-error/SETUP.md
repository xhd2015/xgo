# Scenario

**Feature**: `FetchHTML` failure is the `List` error

```
FetchHTML -> error "html down" -> List -> that error, no versions
```

## Preconditions

- `List` must not invent versions when the fetch fails.

## Steps

1. Inject `FetchHTML` that returns `fmt.Errorf("html down")`.

```go
import (
	"context"
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FetchHTML = func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("html down")
	}
	return nil
}
```
