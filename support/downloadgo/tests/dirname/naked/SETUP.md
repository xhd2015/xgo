# Scenario

**Feature**: `DirName` prefixes a naked version with `go`

```
DirName("1.19.13") -> "go1.19.13"
```

## Preconditions

- Input has no `go` prefix.

## Steps

1. Set `req.Version = "1.19.13"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Version = testVersionNaked
	return nil
}
```
