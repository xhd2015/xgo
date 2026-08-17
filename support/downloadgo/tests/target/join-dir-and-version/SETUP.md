# Scenario

**Feature**: `Target` joins a parent directory with `go1.19.13`

```
Target("/tmp/installed", "go1.19.13") -> "/tmp/installed/go1.19.13"
```

## Preconditions

- `Dir` is the absolute parent `/tmp/installed` (path math only; the dir need
  not exist).
- Version uses the `go` prefix; dest name is still `go1.19.13`.

## Steps

1. Set `req.Dir = "/tmp/installed"` and `req.Version = "go1.19.13"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Dir = "/tmp/installed"
	req.Version = testVersionGo
	return nil
}
```
