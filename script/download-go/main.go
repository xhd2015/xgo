package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	cmd_exec "github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/osinfo"
)

// usage: list, download go1.20.0

const downloadListURL = "https://go.dev/dl"

// download link examples:
//   https://go.dev/dl/go1.22.1.linux-386.tar.gz
//   https://go.dev/dl/go1.22.1.windows-amd64.zip
const baseNameTemplate = "go%s.%s-%s%s"
const downloadLinkPrefix = "https://go.dev/dl/"

const goReleaseDir = "go-release"

func main() {
	args := os.Args[1:]
	var cmd string
	var version string
	var targetDir string
	if len(args) > 0 {
		cmd = args[0]
		if cmd == "download" {
			var err error
			version, targetDir, err = parseDownloadArgs(args[1:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		} else if strings.HasPrefix(cmd, "go") {
			version = cmd
			cmd = "download"
		}
	}
	if cmd == "" {
		fmt.Fprintf(os.Stderr, "requires cmd\n")
		os.Exit(1)
	}

	ctx := context.Background()
	if cmd == "list" {
		goVersions, err := getDownloadVersions(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		for _, goVersion := range goVersions {
			fmt.Printf("go%s\n", goVersion)
		}
		return
	}
	if cmd != "download" {
		fmt.Fprintf(os.Stderr, "unrecognized cmd: %s\n", cmd)
		os.Exit(1)
	}
	if targetDir == "" {
		targetDir = goReleaseDir
	}

	err := downloadGo(ctx, version, targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func parseDownloadArgs(args []string) (version string, targetDir string, err error) {
	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--dir" {
			if i+1 >= n {
				err = fmt.Errorf("%s requires arg", arg)
				return
			}
			targetDir = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			err = fmt.Errorf("unrecognized flag: %s", arg)
			return
		}
		version = arg
	}
	return
}

func downloadGo(ctx context.Context, version string, downloadDir string) error {
	if version == "" {
		return fmt.Errorf("download requires version")
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// goos, err := cmdOutput("go", "env", "GOOS")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v\n", err)
	// 	os.Exit(1)
	// }
	// goarch, err := cmdOutput("go", "env", "GOARCH")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v\n", err)
	// 	os.Exit(1)
	// }

	if goos == "" {
		return fmt.Errorf("requires GOOS")
	}
	if goarch == "" {
		return fmt.Errorf("requires GOARCH")
	}
	nakedVersion := version
	if strings.HasPrefix(version, "go") {
		nakedVersion = strings.TrimPrefix(version, "go")
	}
	err := os.MkdirAll(downloadDir, 0755)
	if err != nil {
		return err
	}

	baseName := fmt.Sprintf(baseNameTemplate, nakedVersion, goos, goarch, getArchiveSuffix(goos))
	downloadLink := downloadLinkPrefix + baseName

	goDirName := filepath.Join(downloadDir, "go"+nakedVersion)

	_, statErr := os.Stat(goDirName)
	if statErr == nil || !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("%s already exists", goDirName)
	}

	goVersions, goVersionErr := getDownloadVersions(ctx)
	if goVersionErr != nil {
		fmt.Fprintf(os.Stderr, "WARNING cannot get go version list:%v\n", goVersionErr)
	}
	var found bool
	for _, goVersion := range goVersions {
		if goVersion == nakedVersion {
			found = true
			break
		}
	}
	if goVersionErr == nil && !found {
		return fmt.Errorf("go%s not found", nakedVersion)
	}

	fmt.Fprintf(os.Stdout, "download from %s\n", downloadLink)
	downloadFile := filepath.Join(downloadDir, baseName)
	err = curlDownload(downloadLink, downloadFile)
	if err != nil {
		return err
	}

	goTmpDir, err := os.MkdirTemp(downloadDir, "go-extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(goTmpDir)
	err = extractArchive(downloadFile, goTmpDir)
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(goTmpDir, "go"), goDirName)
	if err != nil {
		return err
	}
	return nil
}
func curlDownload(url string, file string) error {
	return cmd_exec.Run("curl"+osinfo.EXE_SUFFIX, "-L", "-o", file, url)
}

func getArchiveSuffix(goos string) string {
	if goos == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

func extractArchive(archiveFile string, targetDir string) error {
	if strings.HasSuffix(archiveFile, ".zip") {
		return unzip(archiveFile, targetDir)
	}
	return cmd_exec.Run("tar", "-C", targetDir, "-xzf", archiveFile)
}

func unzip(zipFile string, targetDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := filepath.Clean(f.Name)
		fullPath := filepath.Join(targetDir, name)
		rel, err := filepath.Rel(targetDir, fullPath)
		if err != nil {
			return err
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("zip entry escapes target dir: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fullPath, 0755)
			if err != nil {
				return err
			}
			continue
		}

		err = os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		perm := f.Mode().Perm()
		if perm == 0 {
			perm = 0644
		}
		w, err := os.OpenFile(fullPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
		if err != nil {
			rc.Close()
			return err
		}

		_, copyErr := io.Copy(w, rc)
		closeErr := w.Close()
		rcErr := rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if rcErr != nil {
			return rcErr
		}
	}
	return nil
}

type DownloadInfo struct {
	GoVersion   string // example: 1.18
	DownloadURL string
}

func getDownloadVersions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadListURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("statusCode=%d, resp=%s", resp.StatusCode, bodyData)
	}
	return parseDownloadVersions(string(bodyData)), nil
}

func parseDownloadVersions(htmlContent string) []string {
	var goVersions []string
	// find all div like id="go1.22.1"
	lines := strings.Split(htmlContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "<div ") {
			continue
		}
		const idGo = `id="go`
		idx := strings.Index(line, idGo)
		if idx < 0 {
			continue
		}
		base := idx + len(idGo)
		qIdx := strings.Index(line[base:], `"`)
		if qIdx < 0 {
			continue
		}
		goVersion := line[base : base+qIdx]
		if goVersion == "" {
			continue
		}
		goVersions = append(goVersions, goVersion)
	}
	return goVersions
}
