package downloadgo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDownload_EmptyVersion(t *testing.T) {
	t.Parallel()
	list, get, extract := panicHooks()
	var stdout, stderr bytes.Buffer
	goroot, err := Download(context.Background(), "", Options{
		Dir:          t.TempDir(),
		Stdout:       &stdout,
		Stderr:       &stderr,
		ListVersions: list,
		GetFile:      get,
		Extract:      extract,
	})
	assertLibErrContains(t, err, "download requires version")
	if goroot != "" {
		t.Fatalf("Goroot = %q, want empty on invalid version", goroot)
	}
	assertNoDownloadFrom(t, stdout.String())
}

func TestDownload_EmptyDir(t *testing.T) {
	t.Parallel()
	list, get, extract := panicHooks()
	var stdout, stderr bytes.Buffer
	goroot, err := Download(context.Background(), testVersionNaked, Options{
		Dir:          "",
		Stdout:       &stdout,
		Stderr:       &stderr,
		ListVersions: list,
		GetFile:      get,
		Extract:      extract,
	})
	if err == nil {
		t.Fatal("expected error for empty Dir")
	}
	if goroot != "" {
		t.Fatalf("Goroot = %q, want empty on empty Dir", goroot)
	}
	assertNoDownloadFrom(t, stdout.String())
}

func TestDownload_DestPresentAsDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := wantGoroot(dir, testVersionGo)
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatal(err)
	}
	sentinelPath := filepath.Join(dest, sentinelName)
	if err := os.WriteFile(sentinelPath, []byte(sentinelContents), 0644); err != nil {
		t.Fatal(err)
	}
	list, get, extract := panicHooks()
	var stdout, stderr bytes.Buffer
	goroot, err := Download(context.Background(), testVersionGo, Options{
		Dir:          dir,
		Stdout:       &stdout,
		Stderr:       &stderr,
		ListVersions: list,
		GetFile:      get,
		Extract:      extract,
	})
	if err != nil {
		t.Fatalf("unexpected library error: %v", err)
	}
	if goroot != dest {
		t.Fatalf("Goroot = %q, want %q", goroot, dest)
	}
	assertEmptyWriters(t, stdout.String(), stderr.String())
	assertSentinelUntouched(t, sentinelPath, sentinelContents)
}

func TestDownload_DestPresentAsFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := wantGoroot(dir, testVersionGo)
	if err := os.WriteFile(dest, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}
	list, get, extract := panicHooks()
	var stdout, stderr bytes.Buffer
	goroot, err := Download(context.Background(), testVersionGo, Options{
		Dir:          dir,
		Stdout:       &stdout,
		Stderr:       &stderr,
		ListVersions: list,
		GetFile:      get,
		Extract:      extract,
	})
	if err == nil {
		t.Fatal("expected error when dest exists as a file")
	}
	if goroot != "" {
		t.Fatalf("Goroot = %q, dest-as-file must not be treated as installed", goroot)
	}
	assertNoDownloadFrom(t, stdout.String())
	fi, statErr := os.Stat(dest)
	if statErr != nil {
		t.Fatalf("stat dest: %v", statErr)
	}
	if fi.IsDir() {
		t.Fatal("dest was replaced by a directory")
	}
}

func TestDownload_DestAbsentVersionListedLinuxTarGz(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	log := &hookLog{}
	get, extract := recordSuccessIO(log)
	var stdout, stderr bytes.Buffer
	goroot, err := Download(context.Background(), testVersionNaked, Options{
		Dir:          dir,
		GOOS:         "linux",
		GOARCH:       "amd64",
		Stdout:       &stdout,
		Stderr:       &stderr,
		ListVersions: recordListVersions(log, []string{testVersionNaked}, nil),
		GetFile:      get,
		Extract:      extract,
	})
	if err != nil {
		t.Fatalf("unexpected library error: %v", err)
	}
	want := wantGoroot(dir, testVersionNaked)
	if goroot != want {
		t.Fatalf("Goroot = %q, want %q", goroot, want)
	}
	assertMarkerAt(t, goroot)
	assertFetchRecorded(t, log, dir, testVersionNaked, "linux", "amd64")
	assertStdoutDownloadFrom(t, stdout.String(), archiveURL(testVersionNaked, "linux", "amd64"))
	if stderr.String() != "" {
		t.Fatalf("stderr want empty, got %q", stderr.String())
	}
}

func TestDownload_DestAbsentVersionListedWindowsZip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	log := &hookLog{}
	get, extract := recordSuccessIO(log)
	var stdout, stderr bytes.Buffer
	goroot, err := Download(context.Background(), testVersionGo, Options{
		Dir:          dir,
		GOOS:         "windows",
		GOARCH:       "amd64",
		Stdout:       &stdout,
		Stderr:       &stderr,
		ListVersions: recordListVersions(log, []string{testVersionNaked}, nil),
		GetFile:      get,
		Extract:      extract,
	})
	if err != nil {
		t.Fatalf("unexpected library error: %v", err)
	}
	want := wantGoroot(dir, testVersionGo)
	if goroot != want {
		t.Fatalf("Goroot = %q, want %q", goroot, want)
	}
	assertMarkerAt(t, goroot)
	assertFetchRecorded(t, log, dir, testVersionNaked, "windows", "amd64")
	assertStdoutDownloadFrom(t, stdout.String(), archiveURL(testVersionNaked, "windows", "amd64"))
	if stderr.String() != "" {
		t.Fatalf("stderr want empty, got %q", stderr.String())
	}
}

func TestDownload_DestAbsentVersionAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	log := &hookLog{}
	get, extract := panicGetFileExtract()
	var stdout, stderr bytes.Buffer
	goroot, err := Download(context.Background(), testVersionNaked, Options{
		Dir:          dir,
		GOOS:         "linux",
		GOARCH:       "amd64",
		Stdout:       &stdout,
		Stderr:       &stderr,
		ListVersions: recordListVersions(log, []string{"1.22.1", "1.20.0"}, nil),
		GetFile:      get,
		Extract:      extract,
	})
	assertLibErrContains(t, err, "go1.19.13 not found")
	if goroot != "" {
		t.Fatalf("Goroot = %q, want empty when version is not listed", goroot)
	}
	assertNoDownloadFrom(t, stdout.String())
	if log.listCalls != 1 {
		t.Fatalf("ListVersions calls = %d, want 1", log.listCalls)
	}
	dest := wantGoroot(dir, testVersionNaked)
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("dest %s: %v (want not exist)", dest, statErr)
	}
}

func TestDownload_DestAbsentListWarnProceed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	log := &hookLog{}
	get, extract := recordSuccessIO(log)
	var stdout, stderr bytes.Buffer
	goroot, err := Download(context.Background(), testVersionNaked, Options{
		Dir:          dir,
		GOOS:         "linux",
		GOARCH:       "amd64",
		Stdout:       &stdout,
		Stderr:       &stderr,
		ListVersions: recordListVersions(log, nil, fmt.Errorf("list down")),
		GetFile:      get,
		Extract:      extract,
	})
	if err != nil {
		t.Fatalf("unexpected library error: %v", err)
	}
	want := wantGoroot(dir, testVersionNaked)
	if goroot != want {
		t.Fatalf("Goroot = %q, want %q", goroot, want)
	}
	assertMarkerAt(t, goroot)
	assertFetchRecorded(t, log, dir, testVersionNaked, "linux", "amd64")
	assertStdoutDownloadFrom(t, stdout.String(), archiveURL(testVersionNaked, "linux", "amd64"))
	assertStderrListWarning(t, stderr.String(), "list down")
}
