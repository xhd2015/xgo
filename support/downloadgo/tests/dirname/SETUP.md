# Scenario

**Feature**: `DirName` normalizes a version string to the SDK directory name

```
Caller -> DirName(version) -> "go" + naked version
```

## Preconditions

- `req.Op` is `dirname`. `Dir`, hooks, and writers are unused.
- Both `go1.19.13` and `1.19.13` must produce `go1.19.13`.

## Steps

1. Set `req.Op = "dirname"`.
2. Leaf sets `req.Version` to the spelling under test.
3. `Run` calls `downloadgo.DirName`.
4. Assert `resp.Name` is `go1.19.13`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "dirname"
	return nil
}
```
