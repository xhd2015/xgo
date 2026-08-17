# Scenario

**Feature**: a list error is a warning; download still proceeds

```
ListVersions -> error "list down"
  -> Stderr: WARNING cannot get go version list:list down
  -> GetFile + Extract still run
  -> $Dir/go1.19.13
```

## Preconditions

- List failure is non-fatal. Only a successful list that omits the version
  is `go%s not found`.
- GOOS/GOARCH forced to linux/amd64 for a stable URL.

## Steps

1. Inject `ListVersions` returning `fmt.Errorf("list down")`.
2. Install recording GetFile + Extract.
3. Set `GOOS=linux`, `GOARCH=amd64`.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.GOOS = "linux"
	req.GOARCH = "amd64"
	installListVersions(req, nil, fmt.Errorf("list down"))
	installSuccessIO(req)
	return nil
}
```
