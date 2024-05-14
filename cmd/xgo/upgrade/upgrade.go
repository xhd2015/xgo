package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"time"

	"github.com/xhd2015/xgo/support/github"
	"github.com/xhd2015/xgo/support/httputil"
	"github.com/xhd2015/xgo/support/strutil"
)

const latestURL = "https://github.com/xhd2015/xgo/releases/latest"

func Upgrade(installDir string) error {
	ctx := context.Background()
	if true {
		curXgoVersion, err := cmdXgoVersion()
		if err != nil {
			return err
		}
		// always run a simple go install command
		err = cmdInstallXgo()
		if err != nil {
			return err
		}
		xgoVersionAfterUpdate, err := cmdXgoVersion()
		if err != nil {
			return err
		}
		if xgoVersionAfterUpdate == "" {
			fmt.Fprintf(os.Stderr, "command 'xgo' not found, you may need to add $GOPATH/bin to your PATH\n")
			return nil
		}
		if curXgoVersion == xgoVersionAfterUpdate {
			fmt.Printf("upgraded xgo v%s\n", xgoVersionAfterUpdate)
			return nil
		}
		fmt.Printf("upgraded xgo v%s -> v%s\n", curXgoVersion, xgoVersionAfterUpdate)
		return nil
	}

	fmt.Printf("checking latest version...\n")
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

	curXgoVersion, err := cmdXgoVersion()
	if err != nil {
		return err
	}
	if curXgoVersion == "" {
		fmt.Fprintf(os.Stderr, "command 'xgo' not found on PATH, you may need to add ~/.xgo/bin to your PATH\n")
	}
	if curXgoVersion != "" && curXgoVersion == latestVersion {
		fmt.Printf("congrates, xgo v%s is update to date.\n", curXgoVersion)
		return nil
	}

	file := fmt.Sprintf("xgo%s-%s-%s.tar.gz", latestVersion, goos, goarch)
	targetFile := filepath.Join(tmpDir, file)

	fmt.Printf("downloading xgo v%s...\n", latestVersion)
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
		// Windows
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
			err = os.Rename(filepath.Join(tmpUnzip, name), filepath.Join(installDir, targetName))
			if err != nil {
				return err
			}
		}
	}
	if curXgoVersion == "" {
		fmt.Printf("upgraded xgo v%s\n", latestVersion)
		return nil
	}
	upgradedXgoVersion, err := cmdXgoVersion()
	if err != nil {
		return err
	}
	if upgradedXgoVersion != latestVersion {
		if upgradedXgoVersion == curXgoVersion {
			return fmt.Errorf("WARNING: upgrade xgo v%s -> v%s seems not working, please file a bug", curXgoVersion, latestVersion)
		}
		return fmt.Errorf("WARNING: upgrade xgo v%s -> v%s seems not working, actual version: %s, please file a bug", curXgoVersion, latestVersion, upgradedXgoVersion)
	}
	if curXgoVersion == latestVersion {
		fmt.Printf("upgraded xgo v%s\n", latestVersion)
		return nil
	}
	fmt.Printf("upgraded xgo v%s -> v%s\n", curXgoVersion, latestVersion)

	return nil
}

// if xgo not found, return can be empty
func cmdXgoVersion() (string, error) {
	version, err := cmdOutput("xgo", "version")
	if err != nil && !errors.Is(err, exec.ErrNotFound) {
		return "", err
	}
	return version, nil
}
func cmdInstallXgo() error {
	cmd := exec.Command("go", "install", "github.com/xhd2015/xgo/cmd/xgo@latest")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func cmdOutput(cmd string, args ...string) (string, error) {
	exeCmd := exec.Command(cmd, args...)
	out, err := exeCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

func GetLatestVersion(ctx context.Context, timeout time.Duration, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	releaseTag, _ := github.GetLatestReleaseTag(ctx, url)
	if releaseTag != "" {
		return strings.TrimPrefix(releaseTag, "v"), nil
	}

	// old
	path, err := httputil.Get302Location(ctx, url)
	if err != nil {
		return "", err
	}
	path, ok := strutil.TrimBefore(path, "/xgo-v")
	if !ok {
		path, ok = strutil.TrimBefore(path, "/tag/v")
	}
	if !ok || path == "" {
		return "", fmt.Errorf("expect tag format: xgo-v1.x.x or tag/v1.x.x, actual: %s", path)
	}
	versionName := path
	return versionName, nil
}

func DownloadFile(ctx context.Context, timeout time.Duration, downloadURL string, targetFile string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return httputil.DownloadFile(ctx, downloadURL, targetFile)
}
