# Scenario

**Feature**: windows archive URL uses `.zip`; `go` prefix is accepted on fetch

```
Download("go1.19.13", GOOS=windows, GOARCH=amd64)
  -> https://go.dev/dl/go1.19.13.windows-amd64.zip
```

## Preconditions

- Version uses the `go` prefix so fetch accepts both spellings.
- Suffix is `.zip` only because `GOOS=windows`, not because of host OS.

## Steps

1. Set `req.Version = "go1.19.13"`, `req.GOOS = "windows"`, `req.GOARCH = "amd64"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Version = testVersionGo
	req.GOOS = "windows"
	req.GOARCH = "amd64"
	return nil
}
```
