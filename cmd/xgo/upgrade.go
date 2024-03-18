package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
)

const latestURL = "https://github.com/xhd2015/xgo/releases/latest"

func upgrade(args []string) error {
	var installDir string
	nArg := len(args)
	for i := 0; i < nArg; i++ {
		arg := args[i]
		if arg == "--install-dir" {
			installDir = args[i+1]
			i++
			continue
		}
	}
	ctx := context.Background()
	latestVersion, err := getLatestVersion(ctx, 60*time.Second, latestURL)
	if err != nil {
		return err
	}
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos == "" {
		return fmt.Errorf("requires GOOS")
	}
	if goarch == "" {
		return fmt.Errorf("requires GOARCH")
	}
	tmpDir, err := os.MkdirTemp("", "xgo-download")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	file := fmt.Sprintf("xgo%s-%s-%s.tar.gz", latestVersion, goos, goarch)
	targetFile := filepath.Join(tmpDir, file)

	downloadURL := fmt.Sprintf("%s/download/%s", latestURL, file)
	err = downloadFile(ctx, 5*time.Minute, downloadURL, targetFile)
	if err != nil {
		return fmt.Errorf("download %s: %w", file, err)
	}
	if installDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		installDir = filepath.Join(homeDir, ".xgo", "bin")

		err = os.MkdirAll(installDir, 0755)
		if err != nil {
			return err
		}
	}

	err = unzipBinaries(targetFile, installDir)
	if err != nil {
		return err
	}

	return nil
}

func unzipBinaries(file string, binDir string) error {
	err := cmd.Dir(binDir).Run("tar", "-xzf", file)
	if err != nil {
		return err
	}
	return nil
}

func downloadFile(ctx context.Context, timeout time.Duration, downloadURL string, targetFile string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return fmt.Errorf("%s not exists", downloadURL)
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed: %s", data)
	}
	file, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func getLatestVersion(ctx context.Context, timeout time.Duration, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 302 {
		return "", fmt.Errorf("expect 302 from %s", url)
	}

	loc, err := resp.Location()
	if err != nil {
		return "", err
	}

	path := loc.Path
	path, ok := trimLast(path, "/xgo-v")
	if !ok {
		path, ok = trimLast(path, "/tag/v")
	}
	if !ok || path == "" {
		return "", fmt.Errorf("expect tag format: xgo-v1.x.x or tag/v1.x.x, actual: %s", loc.Path)
	}
	versionName := path
	return versionName, nil
}

func trimLast(s string, p string) (string, bool) {
	i := strings.LastIndex(s, p)
	if i < 0 {
		return s, false
	}
	return s[i+len(p):], true
}
