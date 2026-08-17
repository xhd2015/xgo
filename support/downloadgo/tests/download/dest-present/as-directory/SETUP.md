# Scenario

**Feature**: an existing dest **directory** is treated as already installed

```
$Dir/go1.19.13/SENTINEL exists
  -> Download("go1.19.13")
  -> that path, no error
# ListVersions / GetFile / Extract panic if called
```

## Preconditions

- Dest is a directory (not a file) at `$Dir/go1.19.13` with sentinel
  `SENTINEL` = `keep-me`.
- This is the new library contract (the CLI today errors `already exists`).

## Steps

1. `MkdirAll` the dest dir.
2. Write the sentinel file; record path and contents on `req`.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dest := wantGoroot(req.Dir, req.Version)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	req.SentinelPath = filepath.Join(dest, sentinelName)
	req.SentinelData = sentinelContents
	return os.WriteFile(req.SentinelPath, []byte(req.SentinelData), 0644)
}
```
