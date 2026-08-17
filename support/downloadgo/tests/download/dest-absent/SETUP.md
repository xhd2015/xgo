# Scenario

**Feature**: dest is missing — resolve platform, list, then GetFile + Extract

```
missing $Dir/go{naked}
  -> ListVersions
  -> GetFile(url, archive) + Extract(archive, temp) + rename temp/go -> Target
```

## Preconditions

- `Dir` is a fresh `t.TempDir()`; Target does not exist yet.
- Default version is naked `1.19.13` (dest name still `go1.19.13`).
- Children inject list + I/O hooks. No real archive is unpacked.

## Steps

1. Set `req.Dir = t.TempDir()` and `req.Version = "1.19.13"`.
2. Leaf / sub-branch installs list and I/O hooks.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Dir = t.TempDir()
	req.Version = testVersionNaked
	return nil
}
```
