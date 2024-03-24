package upgrade

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const latestURL = "https://github.com/xhd2015/xgo/releases/latest"

func Upgrade(installDir string) error {
	ctx := context.Background()
	latestVersion, err := GetLatestVersion(ctx, 60*time.Second, latestURL)
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
	err = DownloadFile(ctx, 5*time.Minute, downloadURL, targetFile)
	if err != nil {
		return fmt.Errorf("download %s: %w", file, err)
	}
	if installDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		installDir = filepath.Join(homeDir, ".xgo", "bin")
	}
	err = os.MkdirAll(installDir, 0755)
	if err != nil {
		return err
	}

	if goos != "windows" {
		err = UntargzCmd(targetFile, installDir)
		if err != nil {
			return err
		}
	} else {
		// windows
		tmpUnzip := filepath.Join(tmpDir, "unzip")
		err = os.MkdirAll(tmpUnzip, 0755)
		if err != nil {
			return err
		}

		err = ExtractTarGzFile(targetFile, tmpUnzip)
		if err != nil {
			return err
		}
		var files []fs.DirEntry
		files, err = os.ReadDir(tmpUnzip)
		if err != nil {
			return err
		}
		const exeSuffix = ".exe"
		for _, file := range files {
			name := file.Name()
			targetName := name
			if !file.IsDir() && !strings.HasSuffix(name, exeSuffix) {
				targetName += exeSuffix
			}
			err = os.Rename(filepath.Join(tmpUnzip, name), filepath.Join(installDir, name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func GetLatestVersion(ctx context.Context, timeout time.Duration, url string) (string, error) {
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

func DownloadFile(ctx context.Context, timeout time.Duration, downloadURL string, targetFile string) error {
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
