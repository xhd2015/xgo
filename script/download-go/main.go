package main

import (
	"context"
	"errors"
	"fmt"
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

// download link example: https://go.dev/dl/go1.22.1.linux-386.tar.gz
const baseNameTemplate = "go%s.%s-%s.tar.gz"
const downloadLinkPrefix = "https://go.dev/dl/"

const goReleaseDir = "go-release"

func main() {
	args := os.Args[1:]
	var cmd string
	var version string
	if len(args) > 0 {
		cmd = args[0]
		if cmd == "download" {
			if len(args) > 1 {
				version = args[1]
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

	err := downloadGo(cmd, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
func downloadGo(cmd string, version string) error {
	ctx := context.Background()

	if cmd == "list" {
		goVersions, err := getDownloadVersions(ctx)
		if err != nil {
			return err
		}
		for _, goVersion := range goVersions {
			fmt.Printf("go%s\n", goVersion)
		}
	} else if cmd == "download" {
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
		err := os.MkdirAll(goReleaseDir, 0755)
		if err != nil {
			return err
		}

		baseName := fmt.Sprintf(baseNameTemplate, nakedVersion, goos, goarch)
		downloadLink := downloadLinkPrefix + baseName

		goDirName := filepath.Join(goReleaseDir, "go"+nakedVersion)

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
		downloadFile := filepath.Join(goReleaseDir, baseName)
		err = curlDownload(downloadLink, downloadFile)
		if err != nil {
			return err
		}

		goTmpDir, err := os.MkdirTemp(".", "go")
		if err != nil {
			return err
		}
		defer os.RemoveAll(goTmpDir)
		err = cmd_exec.Run("tar", "-C", goTmpDir, "-xzf", downloadFile)
		if err != nil {
			return err
		}
		err = os.Rename(filepath.Join(goTmpDir, "go"), goDirName)
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("unrecognized cmd: %s", cmd)
	}
	return nil
}
func curlDownload(url string, file string) error {
	return cmd_exec.Run("curl"+osinfo.EXE_SUFFIX, "-L", "-o", file, url)
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
