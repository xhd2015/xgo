package downloadgo

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func TestDirName_GoPrefix(t *testing.T) {
	t.Parallel()
	got := DirName(testVersionGo)
	if got != testDirName {
		t.Fatalf("DirName = %q, want %q", got, testDirName)
	}
}

func TestDirName_Naked(t *testing.T) {
	t.Parallel()
	got := DirName(testVersionNaked)
	if got != testDirName {
		t.Fatalf("DirName = %q, want %q", got, testDirName)
	}
}

func TestTarget_JoinDirAndVersion(t *testing.T) {
	t.Parallel()
	dir := "/tmp/installed"
	got := Target(dir, testVersionGo)
	want := filepath.Join(dir, testDirName)
	if got != want {
		t.Fatalf("Target = %q, want %q", got, want)
	}
}

const matchingHTML = `
<html>
<body>
<div class="toggle" id="go1.22.1"></div>
<div id="go1.21.0"></div>
		<div id="go1.20.3"></div>
<div>no version</div>
<div id="something-else">skip</div>
</body>
</html>
`

func TestList_MatchingDivs(t *testing.T) {
	t.Parallel()
	got, err := List(context.Background(), ListOptions{
		FetchHTML: func(ctx context.Context) (string, error) {
			return matchingHTML, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected library error: %v", err)
	}
	assertVersions(t, got, []string{"1.22.1", "1.21.0", "1.20.3"})
}

const noMatchHTML = `
<html><body>
<div>no version</div>
<div id="something-else">skip</div>
<span id="go1.22.1">not a div</span>
<div id="go"></div>
</body></html>
`

func TestList_NoMatchingDivs(t *testing.T) {
	t.Parallel()
	got, err := List(context.Background(), ListOptions{
		FetchHTML: func(ctx context.Context) (string, error) {
			return noMatchHTML, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected library error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("versions = %#v, want empty", got)
	}
}

func TestList_FetchError(t *testing.T) {
	t.Parallel()
	got, err := List(context.Background(), ListOptions{
		FetchHTML: func(ctx context.Context) (string, error) {
			return "", fmt.Errorf("html down")
		},
	})
	assertLibErrContains(t, err, "html down")
	if len(got) != 0 {
		t.Fatalf("versions = %#v, want empty on fetch error", got)
	}
}
