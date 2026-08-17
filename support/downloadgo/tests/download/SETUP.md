# Scenario

**Feature**: `Download` installs a prebuilt SDK under `$Dir/go{naked}`

```
Download(version, Options{Dir, GOOS, GOARCH, writers, hooks})
  -> goroot path | error
  -> Stdout / Stderr via Options writers
```

## Preconditions

- `req.Op` is `download`.
- Every leaf injects hooks so `Download` never uses production HTTP or
  `curl`/`tar`.
- `Dir` is required (empty is an error leaf). Version accepts `go1.19.13` or
  `1.19.13`.

## Steps

1. Set `req.Op = "download"`.
2. Branch Setup chooses dest state (invalid / present / absent).
3. `Run` calls `downloadgo.Download` with `req` hooks and in-memory writers.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "download"
	return nil
}
```
