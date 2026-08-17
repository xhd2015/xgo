# Scenario

**Feature**: `Download` rejects empty version or empty `Dir` before any fetch

```
Download("", dir) -> error "download requires version"
Download(version, "") -> error
# dest is not consulted; hooks must not run
```

## Preconditions

- Input validation is independent of dest state.
- Panicking hooks prove the library does not start a download.

## Steps

1. Install panicking `ListVersions` / `GetFile` / `Extract`.
2. Leaf sets the one empty field (version or `Dir`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	installPanicHooks(req)
	return nil
}
```
