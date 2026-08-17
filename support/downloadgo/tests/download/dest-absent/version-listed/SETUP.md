# Scenario

**Feature**: listed version is fetched and extracted into Target

```
ListVersions -> ["1.19.13"]
  -> GetFile writes dummy archive
  -> Extract creates destDir/go/INSTALLED
  -> Download returns $Dir/go1.19.13
```

## Preconditions

- List includes the naked version. GetFile writes a dummy file; Extract
  creates `destDir/go/INSTALLED` (`ok`). Download must rename that `go/`
  onto Target.
- Children set `GOOS` / `GOARCH` so the constructed URL is deterministic.

## Steps

1. Inject `ListVersions` returning `["1.19.13"]`.
2. Install recording GetFile + Extract.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	installListVersions(req, []string{testVersionNaked}, nil)
	installSuccessIO(req)
	return nil
}
```
