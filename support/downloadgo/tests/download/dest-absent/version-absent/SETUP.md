# Scenario

**Feature**: a successful list that omits the version is a hard not-found

```
ListVersions -> ["1.22.1", "1.20.0"]  # no 1.19.13
  -> error "go1.19.13 not found"
  -> GetFile / Extract unused
```

## Preconditions

- List succeeds. Absence is only an error when the list itself succeeded.
- GetFile / Extract panic if called.

## Steps

1. Inject `ListVersions` returning other versions.
2. Install panicking GetFile + Extract.
3. Set linux/amd64 so a mistaken fetch would still be deterministic.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.GOOS = "linux"
	req.GOARCH = "amd64"
	installListVersions(req, []string{"1.22.1", "1.20.0"}, nil)
	installPanicGetFileExtract(req)
	return nil
}
```
