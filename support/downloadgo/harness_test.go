package downloadgo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testVersionNaked  = "1.19.13"
	testVersionGo     = "go1.19.13"
	testDirName       = "go1.19.13"
	installedMarker   = "INSTALLED"
	installedContents = "ok"
	sentinelName      = "SENTINEL"
	sentinelContents  = "keep-me"
	listWarnPrefix    = "WARNING cannot get go version list:"
)

type hookLog struct {
	listCalls      int
	getFileCalls   int
	extractCalls   int
	getFileURL     string
	getFileDest    string
	extractArchive string
	extractDest    string
}

func goDirName(version string) string {
	if strings.HasPrefix(version, "go") {
		return version
	}
	return "go" + version
}

func wantGoroot(dir, version string) string {
	return filepath.Join(dir, goDirName(version))
}

func archiveURL(naked, goos, goarch string) string {
	suffix := ".tar.gz"
	if goos == "windows" {
		suffix = ".zip"
	}
	return "https://go.dev/dl/go" + naked + "." + goos + "-" + goarch + suffix
}

func archiveBase(naked, goos, goarch string) string {
	suffix := ".tar.gz"
	if goos == "windows" {
		suffix = ".zip"
	}
	return "go" + naked + "." + goos + "-" + goarch + suffix
}

func writeInstalledSDK(destDir string) error {
	goDir := filepath.Join(destDir, "go")
	if err := os.MkdirAll(goDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(goDir, installedMarker), []byte(installedContents), 0644)
}

func panicHooks() (listVersions func(context.Context) ([]string, error), getFile func(context.Context, string, string) error, extract func(string, string) error) {
	listVersions = func(ctx context.Context) ([]string, error) {
		panic("ListVersions must not be called")
	}
	getFile = func(ctx context.Context, url, dest string) error {
		panic("GetFile must not be called")
	}
	extract = func(archiveFile, destDir string) error {
		panic("Extract must not be called")
	}
	return
}

func panicGetFileExtract() (getFile func(context.Context, string, string) error, extract func(string, string) error) {
	getFile = func(ctx context.Context, url, dest string) error {
		panic("GetFile must not be called")
	}
	extract = func(archiveFile, destDir string) error {
		panic("Extract must not be called")
	}
	return
}

func recordListVersions(log *hookLog, versions []string, listErr error) func(context.Context) ([]string, error) {
	return func(ctx context.Context) ([]string, error) {
		log.listCalls++
		if listErr != nil {
			return nil, listErr
		}
		out := make([]string, len(versions))
		copy(out, versions)
		return out, nil
	}
}

func recordSuccessIO(log *hookLog) (getFile func(context.Context, string, string) error, extract func(string, string) error) {
	getFile = func(ctx context.Context, url, dest string) error {
		log.getFileCalls++
		log.getFileURL = url
		log.getFileDest = dest
		return os.WriteFile(dest, []byte("dummy-archive"), 0644)
	}
	extract = func(archiveFile, destDir string) error {
		log.extractCalls++
		log.extractArchive = archiveFile
		log.extractDest = destDir
		return writeInstalledSDK(destDir)
	}
	return
}

func assertLibErrContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected library error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("library error %q, want substring %q", err.Error(), substr)
	}
}

func assertVersions(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("versions = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("versions[%d] = %q, want %q (full %#v)", i, got[i], want[i], got)
		}
	}
}

func assertMarkerAt(t *testing.T, goroot string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(goroot, installedMarker))
	if err != nil {
		t.Fatalf("read installed marker: %v", err)
	}
	if string(data) != installedContents {
		t.Fatalf("marker contents %q, want %q", string(data), installedContents)
	}
}

func assertSentinelUntouched(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read sentinel: %v", err)
	}
	if string(data) != want {
		t.Fatalf("sentinel %q, want %q (dest was overwritten)", string(data), want)
	}
}

func assertNoDownloadFrom(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "download from") {
		t.Fatalf("stdout must not contain download from:\n%s", stdout)
	}
}

func assertStdoutDownloadFrom(t *testing.T, stdout, url string) {
	t.Helper()
	want := "download from " + url + "\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func assertStderrListWarning(t *testing.T, stderr, errText string) {
	t.Helper()
	want := listWarnPrefix + errText + "\n"
	if stderr != want {
		t.Fatalf("stderr = %q, want %q", stderr, want)
	}
}

func assertEmptyWriters(t *testing.T, stdout, stderr string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout want empty, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr want empty, got %q", stderr)
	}
}

func assertFetchRecorded(t *testing.T, log *hookLog, dir, naked, goos, goarch string) {
	t.Helper()
	wantURL := archiveURL(naked, goos, goarch)
	wantDest := filepath.Join(dir, archiveBase(naked, goos, goarch))
	if log.listCalls != 1 {
		t.Fatalf("ListVersions calls = %d, want 1", log.listCalls)
	}
	if log.getFileCalls != 1 {
		t.Fatalf("GetFile calls = %d, want 1", log.getFileCalls)
	}
	if log.extractCalls != 1 {
		t.Fatalf("Extract calls = %d, want 1", log.extractCalls)
	}
	if log.getFileURL != wantURL {
		t.Fatalf("GetFile url = %q, want %q", log.getFileURL, wantURL)
	}
	if log.getFileDest != wantDest {
		t.Fatalf("GetFile dest = %q, want %q", log.getFileDest, wantDest)
	}
	if log.extractArchive != wantDest {
		t.Fatalf("Extract archive = %q, want %q", log.extractArchive, wantDest)
	}
}
