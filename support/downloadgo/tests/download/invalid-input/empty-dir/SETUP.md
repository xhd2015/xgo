# Scenario

**Feature**: empty `Options.Dir` is rejected

```
Download("1.19.13", Options{Dir: ""}) -> error
```

## Preconditions

- Version is valid (`1.19.13`) so the only missing input is `Dir`.
- Hooks panic if called. Cwd is undetermined; the library must not fall back
  to a relative install.

## Steps

1. Set `req.Version = "1.19.13"`.
2. Set `req.Dir = ""`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Version = testVersionNaked
	req.Dir = ""
	return nil
}
```
