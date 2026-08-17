package downloadgo

import (
	"testing"
)

func TestParseDownloadVersions(t *testing.T) {
	html := `
		<html>
		<body>
		<div id="go1.22.1"></div>
		<div id="go1.21.0"></div>
		<div id="go1.20.3"></div>
		<div>no version</div>
		<div id="something-else">skip</div>
		</body>
		</html>
	`

	versions := parseDownloadVersions(html)

	expected := []string{"1.22.1", "1.21.0", "1.20.3"}
	if len(versions) != len(expected) {
		t.Fatalf("got %d versions, want %d: %v", len(versions), len(expected), versions)
	}
	for i, v := range versions {
		if v != expected[i] {
			t.Errorf("version[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestParseDownloadVersionsEmpty(t *testing.T) {
	versions := parseDownloadVersions("<html><body></body></html>")
	if len(versions) != 0 {
		t.Errorf("expected empty, got %v", versions)
	}
}
