# Scenario

**Feature**: `DirName` keeps an already-prefixed version

```
DirName("go1.19.13") -> "go1.19.13"
```

## Preconditions

- Input already starts with `go`.

## Steps

1. Set `req.Version = "go1.19.13"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Version = testVersionGo
	return nil
}
```
