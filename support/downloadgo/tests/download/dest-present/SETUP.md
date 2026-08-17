# Scenario

**Feature**: Target already exists on disk — short-circuit before fetch

```
stat($Dir/go1.19.13)
  -> directory -> return that path (no hooks, no "download from")
  -> file      -> error (not treated as installed)
```

## Preconditions

- Version is `go1.19.13`; dest name is `go1.19.13`.
- `Dir` is a fresh `t.TempDir()`.
- Panicking hooks: if the implementer skips the short-circuit, the test
  panics instead of hitting the network.

## Steps

1. Set `req.Dir = t.TempDir()` and `req.Version = "go1.19.13"`.
2. Install panicking hooks.
3. Leaf creates dest as a directory (with sentinel) or as a file.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Dir = t.TempDir()
	req.Version = testVersionGo
	installPanicHooks(req)
	return nil
}
```
