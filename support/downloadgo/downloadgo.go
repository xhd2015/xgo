package downloadgo

import (
	"context"
	"io"
	"path/filepath"
	"strings"
)

const (
	downloadListURL    = "https://go.dev/dl"
	downloadLinkPrefix = "https://go.dev/dl/"
	baseNameTemplate   = "go%s.%s-%s%s"
)

// Options controls Download. Nil hooks use production implementations
// (HTTP list, curl GetFile, tar/zip Extract). Writers must not be process
// stdio unless the caller passes them explicitly.
type Options struct {
	Dir, GOOS, GOARCH string
	Stdout, Stderr    io.Writer
	ListVersions      func(ctx context.Context) ([]string, error)
	GetFile           func(ctx context.Context, url, dest string) error
	Extract           func(archiveFile, destDir string) error
}

// ListOptions controls List. A nil FetchHTML GETs https://go.dev/dl.
type ListOptions struct {
	FetchHTML func(ctx context.Context) (string, error)
}

// DirName maps a version spelling to the SDK directory name.
// Both "go1.19.13" and "1.19.13" become "go1.19.13".
func DirName(version string) string {
	if strings.HasPrefix(version, "go") {
		return version
	}
	return "go" + version
}

// Target is filepath.Join(dir, DirName(version)).
func Target(dir, version string) string {
	return filepath.Join(dir, DirName(version))
}

func nakedVersion(version string) string {
	return strings.TrimPrefix(version, "go")
}

func writerOrDiscard(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}
	return w
}
