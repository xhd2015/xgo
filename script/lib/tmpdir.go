package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func TmpDir() string {
	dir := os.TempDir()
	if dir == "" {
		return "/tmp"
	}
	return dir
}

var xgoPatterns = []string{
	"xgo-test-fb-",
	"xgo-test-prog-",
	"xgo-test-repeat-",
	"xgo-compare-",
	"xgo-patch-test-",
	"xgo-wt-",
}

type DirEntry struct {
	Name string
	Size int64
}

func ListXgoTempDirs() []DirEntry {
	var result []DirEntry
	dir := TmpDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		for _, p := range xgoPatterns {
			if strings.HasPrefix(name, p) {
				path := filepath.Join(dir, name)
				size := dirSize(path)
				result = append(result, DirEntry{Name: name, Size: size})
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

type PatternStat struct {
	Pattern string
	Count   int
	Size    int64
}

func StatsByPattern() []PatternStat {
	entries := ListXgoTempDirs()
	stats := make(map[string]*PatternStat)
	var patterns []string

	for _, e := range entries {
		p := patternFor(e.Name)
		s, ok := stats[p]
		if !ok {
			s = &PatternStat{Pattern: p}
			stats[p] = s
			patterns = append(patterns, p)
		}
		s.Count++
		s.Size += e.Size
	}
	sort.Strings(patterns)
	var result []PatternStat
	for _, p := range patterns {
		result = append(result, *stats[p])
	}
	return result
}

func patternFor(name string) string {
	for _, p := range xgoPatterns {
		if strings.HasPrefix(name, p) {
			return p + "*"
		}
	}
	return name
}

func CleanupXgoTempDirs() (removed int, freedBytes int64) {
	dir := TmpDir()
	for _, entry := range ListXgoTempDirs() {
		path := filepath.Join(dir, entry.Name)
		os.RemoveAll(path)
		removed++
		freedBytes += entry.Size
	}
	return
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		size += info.Size()
		return nil
	})
	return size
}

func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
