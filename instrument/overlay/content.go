package overlay

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// contentDirBytes is the number of bytes of the sha256 of a file content
// that name the directory it is written to (as hex).
const contentDirBytes = 8

// WriteFileIfChanged writes data to file unless it already has that
// content, through a temporary file renamed into place: readers see the
// old or the new content, never a partial one.
func WriteFileIfChanged(file string, data []byte) error {
	if old, err := os.ReadFile(file); err == nil && bytes.Equal(old, data) {
		return nil
	}
	return writeFileAtomic(file, data)
}

// writeFileOnce writes data to file unless the file exists: file is named
// after data, so an existing one has the same content. It is written to a
// temporary file renamed into place, so concurrent readers and writers
// see either no file or the whole of it. An existing file is touched.
func writeFileOnce(file string, data []byte) error {
	now := time.Now()
	if err := os.Chtimes(file, now, now); err == nil {
		return nil
	}
	err := writeFileAtomic(file, data)
	if err != nil {
		// on windows the rename fails if another run renamed the same
		// content into place first
		if _, statErr := os.Stat(file); statErr == nil {
			return nil
		}
	}
	return err
}

func writeFileAtomic(file string, data []byte) error {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(file)+"-*")
	if err != nil {
		return err
	}
	_, err = tmp.Write(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Chmod(tmp.Name(), 0644)
	}
	if err == nil {
		err = os.Rename(tmp.Name(), file)
	}
	if err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}

// CleanGoOverlay removes the files MakeGoOverlay wrote under overlayDir
// that no run used for maxAge, and their content directories once empty.
// Only files in content directories, and leftover temporary files, are
// considered: other files under overlayDir are not written by
// MakeGoOverlay.
func CleanGoOverlay(overlayDir string, maxAge time.Duration) error {
	deadline := time.Now().Add(-maxAge)
	var dirs []string
	err := filepath.Walk(overlayDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// removed concurrently, or unreadable: skip
			return nil
		}
		if info.IsDir() {
			// an empty one can be one a run just created to write to
			if isContentDir(info.Name()) && info.ModTime().Before(deadline) {
				dirs = append(dirs, path)
			}
			return nil
		}
		if !info.ModTime().Before(deadline) {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".tmp-") || isContentDir(filepath.Base(filepath.Dir(path))) {
			// a run reusing the file touches it: check again just before
			// removing, rather than on what the walk saw
			if cur, err := os.Stat(path); err == nil && cur.ModTime().Before(deadline) {
				os.Remove(path)
			}
		}
		return nil
	})
	// fails on directories that are not empty
	for _, dir := range dirs {
		os.Remove(dir)
	}
	return err
}

func isContentDir(name string) bool {
	if len(name) != 2*contentDirBytes {
		return false
	}
	for _, c := range name {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// WriteContentFile writes data as name in a content directory under
// overlayDir, like MakeGoOverlay does, and returns its path.
func WriteContentFile(overlayDir string, name string, data []byte) (string, error) {
	sum := sha256.Sum256(data)
	file := filepath.Join(overlayDir, hex.EncodeToString(sum[:contentDirBytes]), name)
	if err := writeFileOnce(file, data); err != nil {
		return "", err
	}
	return file, nil
}

// WriteContentFiles writes files, by name, together into a content
// directory under dir named after all of them, like WriteContentFile, and
// returns that directory.
func WriteContentFiles(dir string, files map[string][]byte) (string, error) {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	h := sha256.New()
	for _, name := range names {
		fmt.Fprintf(h, "%s %d\n", name, len(files[name]))
		h.Write(files[name])
	}
	contentDir := filepath.Join(dir, hex.EncodeToString(h.Sum(nil)[:contentDirBytes]))
	for _, name := range names {
		if err := writeFileOnce(filepath.Join(contentDir, name), files[name]); err != nil {
			return "", err
		}
	}
	return contentDir, nil
}

// cleanStampFile is touched under the overlay dir whenever it is cleaned.
const cleanStampFile = ".cleaned"

// MaybeCleanGoOverlay runs CleanGoOverlay with maxAge if it did not run
// within interval. It must run before MakeGoOverlay: a run that finds a
// file to reuse touches it after the cleanup decided on its age.
func MaybeCleanGoOverlay(overlayDir string, interval time.Duration, maxAge time.Duration) error {
	stamp := filepath.Join(overlayDir, cleanStampFile)
	if info, err := os.Stat(stamp); err == nil && time.Since(info.ModTime()) < interval {
		return nil
	}
	if err := os.MkdirAll(overlayDir, 0755); err != nil {
		return err
	}
	// stamped first so that concurrent runs do not all clean
	if err := os.WriteFile(stamp, nil, 0644); err != nil {
		return err
	}
	if err := CleanGoOverlay(overlayDir, maxAge); err != nil {
		os.Remove(stamp)
		return err
	}
	return nil
}
