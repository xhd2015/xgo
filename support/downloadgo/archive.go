package downloadgo

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	cmd_exec "github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/osinfo"
)

func getArchiveSuffix(goos string) string {
	if goos == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

func curlDownload(ctx context.Context, url string, file string) error {
	_ = ctx
	return cmd_exec.Run("curl"+osinfo.EXE_SUFFIX, "-L", "-o", file, url)
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
		if err := unzipFile(f, targetDir); err != nil {
			return err
		}
	}
	return nil
}

func unzipFile(f *zip.File, targetDir string) (err error) {
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
		return os.MkdirAll(fullPath, 0755)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	perm := f.Mode().Perm()
	if perm == 0 {
		perm = 0644
	}
	w, err := os.OpenFile(fullPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer func() {
		cerr := w.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(w, rc); err != nil {
		return err
	}
	return nil
}
