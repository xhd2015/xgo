# Scenario

**Feature**: `Target` joins an install dir with `DirName(version)`

```
Caller -> Target(dir, version) -> filepath.Join(dir, DirName(version))
```

## Preconditions

- `req.Op` is `target`. No filesystem access; `Target` is a pure join.
- Version spelling still goes through `DirName`.

## Steps

1. Set `req.Op = "target"`.
2. Leaf sets `Dir` and `Version`.
3. `Run` calls `downloadgo.Target`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "target"
	return nil
}
```
