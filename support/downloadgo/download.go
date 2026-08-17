package downloadgo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Download installs a prebuilt Go SDK at Target(opts.Dir, version).
// An existing dest directory is treated as already installed.
func Download(ctx context.Context, version string, opts Options) (string, error) {
	if version == "" {
		return "", fmt.Errorf("download requires version")
	}
	if opts.Dir == "" {
		return "", fmt.Errorf("download requires dir")
	}

	dest := Target(opts.Dir, version)
	fi, statErr := os.Stat(dest)
	if statErr == nil {
		if fi.IsDir() {
			return dest, nil
		}
		return "", fmt.Errorf("%s exists and is not a directory", dest)
	}
	if !errors.Is(statErr, os.ErrNotExist) {
		return "", statErr
	}

	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := opts.GOARCH
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	if goos == "" {
		return "", fmt.Errorf("requires GOOS")
	}
	if goarch == "" {
		return "", fmt.Errorf("requires GOARCH")
	}

	stdout := writerOrDiscard(opts.Stdout)
	stderr := writerOrDiscard(opts.Stderr)

	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return "", err
	}

	listVersions := opts.ListVersions
	if listVersions == nil {
		listVersions = func(ctx context.Context) ([]string, error) {
			return List(ctx, ListOptions{})
		}
	}
	goVersions, goVersionErr := listVersions(ctx)
	if goVersionErr != nil {
		fmt.Fprintf(stderr, "WARNING cannot get go version list:%v\n", goVersionErr)
	} else {
		naked := nakedVersion(version)
		found := false
		for _, goVersion := range goVersions {
			if goVersion == naked {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("go%s not found", naked)
		}
	}

	naked := nakedVersion(version)
	baseName := fmt.Sprintf(baseNameTemplate, naked, goos, goarch, getArchiveSuffix(goos))
	downloadLink := downloadLinkPrefix + baseName
	fmt.Fprintf(stdout, "download from %s\n", downloadLink)

	getFile := opts.GetFile
	if getFile == nil {
		getFile = curlDownload
	}
	downloadFile := filepath.Join(opts.Dir, baseName)
	if err := getFile(ctx, downloadLink, downloadFile); err != nil {
		return "", err
	}

	extract := opts.Extract
	if extract == nil {
		extract = extractArchive
	}
	goTmpDir, err := os.MkdirTemp(opts.Dir, "go-extract-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(goTmpDir)
	if err := extract(downloadFile, goTmpDir); err != nil {
		return "", err
	}
	if err := os.Rename(filepath.Join(goTmpDir, "go"), dest); err != nil {
		return "", err
	}
	return dest, nil
}
