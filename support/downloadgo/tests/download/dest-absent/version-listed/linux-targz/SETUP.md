# Scenario

**Feature**: linux archive URL uses `.tar.gz`

```
GOOS=linux GOARCH=amd64
  -> https://go.dev/dl/go1.19.13.linux-amd64.tar.gz
  -> stdout "download from <url>"
```

## Preconditions

- Version is naked `1.19.13`. Platform is forced via opts (not `runtime.GOOS`).

## Steps

1. Set `req.GOOS = "linux"` and `req.GOARCH = "amd64"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.GOOS = "linux"
	req.GOARCH = "amd64"
	return nil
}
```
