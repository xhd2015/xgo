# Scenario

**Feature**: empty version is rejected

```
Download("", Options{Dir: temp}) -> error "download requires version"
```

## Preconditions

- `Dir` is a real temp directory so the only missing input is version.
- Hooks panic if called.

## Steps

1. Set `req.Version = ""`.
2. Set `req.Dir` to `t.TempDir()`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Version = ""
	req.Dir = t.TempDir()
	return nil
}
```
