package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetArchiveSuffix(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{"windows", ".zip"},
		{"linux", ".tar.gz"},
		{"darwin", ".tar.gz"},
		{"freebsd", ".tar.gz"},
	}

	for _, tt := range tests {
		got := getArchiveSuffix(tt.goos)
		if got != tt.want {
			t.Errorf("getArchiveSuffix(%q) = %q, want %q", tt.goos, got, tt.want)
		}
	}
}

func TestParseDownloadVersions(t *testing.T) {
	html := `
		<html>
		<body>
		<div id="go1.22.1"></div>
		<div id="go1.21.0"></div>
		<div id="go1.20.3"></div>
		<div>no version</div>
		<div id="something-else">skip</div>
		</body>
		</html>
	`

	versions := parseDownloadVersions(html)

	expected := []string{"1.22.1", "1.21.0", "1.20.3"}
	if len(versions) != len(expected) {
		t.Fatalf("got %d versions, want %d: %v", len(versions), len(expected), versions)
	}
	for i, v := range versions {
		if v != expected[i] {
			t.Errorf("version[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestParseDownloadVersionsEmpty(t *testing.T) {
	versions := parseDownloadVersions("<html><body></body></html>")
	if len(versions) != 0 {
		t.Errorf("expected empty, got %v", versions)
	}
}

func TestExtractArchiveTarGz(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tar extraction test requires tar command")
	}

	tmpDir := t.TempDir()

	// Verify that a non-.zip file dispatches to tar command
	err := extractArchive("/nonexistent/test.tar.gz", tmpDir)
	if err == nil {
		t.Fatal("expected error for nonexistent tar.gz")
	}
}

func TestExtractArchiveZip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid zip for testing
	zipPath := filepath.Join(tmpDir, "test.zip")
	if err := createTestZip(zipPath, map[string]string{
		"go/README": "hello",
	}); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err := extractArchive(zipPath, extractDir)
	if err != nil {
		t.Fatalf("extractArchive failed: %v", err)
	}

	// Verify file was extracted
	content, err := os.ReadFile(filepath.Join(extractDir, "go", "README"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "hello" {
		t.Errorf("expected 'hello', got %q", string(content))
	}
}

func TestUnzip(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "test.zip")
	if err := createTestZip(zipPath, map[string]string{
		"go/bin/go":    "mock-binary",
		"go/src/main":  "package main",
		"go/README.md": "# Go Docs",
	}); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := unzip(zipPath, extractDir); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	// Verify all files extracted
	testCases := []struct {
		path    string
		content string
	}{
		{"go/bin/go", "mock-binary"},
		{"go/src/main", "package main"},
		{"go/README.md", "# Go Docs"},
	}

	for _, tc := range testCases {
		content, err := os.ReadFile(filepath.Join(extractDir, tc.path))
		if err != nil {
			t.Errorf("failed to read %s: %v", tc.path, err)
			continue
		}
		if string(content) != tc.content {
			t.Errorf("%s content = %q, want %q", tc.path, string(content), tc.content)
		}
	}
}

func TestUnzipDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "test.zip")
	if err := createTestZipWithDirs(zipPath, []string{
		"go/",
		"go/bin/",
		"go/src/",
	}, map[string]string{
		"go/bin/go": "binary",
	}); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := unzip(zipPath, extractDir); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	// Check directories exist
	for _, dir := range []string{"go", "go/bin", "go/src"} {
		fi, err := os.Stat(filepath.Join(extractDir, dir))
		if err != nil {
			t.Errorf("directory %s not found: %v", dir, err)
			continue
		}
		if !fi.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestUnzipPathTraversal_RelativeParent(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "test.zip")
	if err := createTestZipWithRawNames(zipPath, map[string]string{
		"../evil": "malicious content",
	}); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err := unzip(zipPath, extractDir)
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "escapes target dir") {
		t.Errorf("expected 'escapes target dir' error, got: %v", err)
	}
}

func TestUnzipPathTraversal_NestedParent(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "test.zip")
	if err := createTestZipWithRawNames(zipPath, map[string]string{
		"foo/../../../evil": "malicious content",
	}); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err := unzip(zipPath, extractDir)
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "escapes target dir") {
		t.Errorf("expected 'escapes target dir' error, got: %v", err)
	}
}

func TestUnzipPathTraversal_SingleDot(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "test.zip")
	if err := createTestZipWithRawNames(zipPath, map[string]string{
		"go/./bin/go": "binary",
	}); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := unzip(zipPath, extractDir); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	// filepath.Clean should normalize . away
	content, err := os.ReadFile(filepath.Join(extractDir, "go", "bin", "go"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "binary" {
		t.Errorf("expected 'binary', got %q", string(content))
	}
}

func TestUnzipFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "test.zip")
	if err := createTestZipWithPerms(zipPath, map[string]zipEntry{
		"go/bin/go": {content: "binary", mode: 0755},
		"go/readme": {content: "text", mode: 0644},
	}); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := unzip(zipPath, extractDir); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	fi, err := os.Stat(filepath.Join(extractDir, "go", "bin", "go"))
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if fi.Mode().Perm() != 0755 {
		t.Errorf("expected 0755, got %04o", fi.Mode().Perm())
	}

	fi2, err := os.Stat(filepath.Join(extractDir, "go", "readme"))
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if fi2.Mode().Perm() != 0644 {
		t.Errorf("expected 0644, got %04o", fi2.Mode().Perm())
	}
}

func TestUnzipEmptyZip(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "test.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	zw := zip.NewWriter(f)
	if err := zw.Close(); err != nil {
		f.Close()
		t.Fatalf("failed to close zip writer: %v", err)
	}
	f.Close()

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := unzip(zipPath, extractDir); err != nil {
		t.Fatalf("unzip empty zip should succeed: %v", err)
	}
}

// Helper functions to create test zip files

func createTestZip(zipPath string, files map[string]string) error {
	entries := make(map[string]zipEntry)
	for name, content := range files {
		entries[name] = zipEntry{content: content}
	}
	return createTestZipWithPerms(zipPath, entries)
}

type zipEntry struct {
	content string
	mode    os.FileMode
}

func createTestZipWithPerms(zipPath string, files map[string]zipEntry) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, entry := range files {
		hdr := &zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		}
		if entry.mode != 0 {
			hdr.SetMode(entry.mode)
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(entry.content)); err != nil {
			return err
		}
	}
	return zw.Close()
}

func createTestZipWithDirs(zipPath string, dirs []string, files map[string]string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	for _, dir := range dirs {
		_, err := zw.CreateHeader(&zip.FileHeader{
			Name: dir,
			Method: zip.Store,
		})
		if err != nil {
			return err
		}
	}

	for name, content := range files {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		})
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(content)); err != nil {
			return err
		}
	}

	return zw.Close()
}

func createTestZipWithRawNames(zipPath string, files map[string]string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		})
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(content)); err != nil {
			return err
		}
	}
	return zw.Close()
}
