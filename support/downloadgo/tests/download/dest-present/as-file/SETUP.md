# Scenario

**Feature**: dest existing as a **file** is not treated as installed

```
$Dir/go1.19.13 is a file
  -> Download -> error
# hooks unused; path stays a file
```

## Preconditions

- Same path the directory case would use, but it is a regular file.

## Steps

1. Write a regular file at `$Dir/go1.19.13`.

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dest := wantGoroot(req.Dir, req.Version)
	return os.WriteFile(dest, []byte("not a directory"), 0644)
}
```
