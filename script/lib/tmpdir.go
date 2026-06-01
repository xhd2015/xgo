package lib

import (
	"fmt"
	"os"
	"os/exec"
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
	"xgo-unified-",
	"xgo-wt-",
}

type DirEntry struct {
	Name string
	Size int64
}

func ListXgoTempDirs() []DirEntry {
	return listXgoTempDirs(true)
}

func ListXgoTempDirsFast() []DirEntry {
	return listXgoTempDirs(false)
}

func listXgoTempDirs(calcSize bool) []DirEntry {
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
				var size int64
				if calcSize {
					size = dirSize(filepath.Join(dir, name))
				}
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
	entries := ListXgoTempDirsFast()
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name)
		Logf("removing: %s", entry.Name)
		os.RemoveAll(path)
		removed++
	}
	return
}

func FindRepoRoot() string {
	if _, err := os.Stat(filepath.Join("cmd", "xgo", "main.go")); err == nil {
		if cwd, err := os.Getwd(); err == nil {
			return cwd
		}
	}
	return ""
}

func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func GitGorootDirs() []string {
	root := FindRepoRoot()
	if root == "" {
		return nil
	}
	gitVersionedDir := filepath.Join(root, "go-release-git-versioned")
	entries, err := os.ReadDir(gitVersionedDir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(gitVersionedDir, e.Name())
		if isGitRepo(path) {
			dirs = append(dirs, path)
		}
	}
	sort.Strings(dirs)
	return dirs
}

type GitGorootStatus struct {
	Dir     string
	Removed int
	Err     error
}

func CleanGitGoroots() []GitGorootStatus {
	dirs := GitGorootDirs()
	var results []GitGorootStatus
	tmpDir := TmpDir()
	for _, dir := range dirs {
		status := GitGorootStatus{Dir: dir}
		Logf("cleaning git goroot: %s", dir)

		out, err := execGit(dir, "reset", "--hard", "HEAD")
		if err != nil {
			status.Err = fmt.Errorf("reset: %w\n%s", err, out)
			results = append(results, status)
			continue
		}

		out, err = execGit(dir, "clean", "-fd")
		if err != nil {
			status.Err = fmt.Errorf("clean: %w\n%s", err, out)
			results = append(results, status)
			continue
		}

		// Prune worktree entries whose directories no longer exist
		pruneOut, err := execGit(dir, "worktree", "prune", "--verbose")
		if pruneOut != "" {
			for _, line := range strings.Split(strings.TrimSpace(pruneOut), "\n") {
				if line != "" {
					Logf("  worktree pruned: %s", line)
				}
			}
		}

		// Force-remove surviving worktrees in temp dir that match xgo patterns
		if err := removeStaleWorktrees(dir, tmpDir); err != nil {
			Logf("  worktree remove warning: %v", err)
		}

		status.Err = err
		results = append(results, status)
	}
	return results
}

func removeStaleWorktrees(gitDir, tmpDir string) error {
	// Resolve real path to handle macOS /var vs /private/var symlink
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		realTmpDir = tmpDir
	}

	listOut, err := execGit(gitDir, "worktree", "list", "--porcelain")
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}
	wts := parseWorktreeList(listOut)
	for _, w := range wts {
		if w.bare {
			continue
		}
		if !strings.HasPrefix(w.path, tmpDir) && !strings.HasPrefix(w.path, realTmpDir) {
			continue
		}
		base := filepath.Base(w.path)
		if !matchesXgoPattern(base) {
			continue
		}
		Logf("  removing worktree: %s", w.path)
		if _, err := execGit(gitDir, "worktree", "remove", "--force", w.path); err != nil {
			Logf("  worktree remove failed (will try fs cleanup): %v", err)
		}
		if err := os.RemoveAll(w.path); err != nil {
			Logf("  remove worktree dir failed: %v", err)
		}
	}
	return nil
}

func parseWorktreeList(out string) []struct {
	path     string
	bare     bool
	detached bool
} {
	var wts []struct {
		path     string
		bare     bool
		detached bool
	}
	lines := strings.Split(out, "\n")
	var cur *struct {
		path     string
		bare     bool
		detached bool
	}
	for _, line := range lines {
		if line == "" {
			if cur != nil && cur.path != "" {
				wts = append(wts, *cur)
			}
			cur = nil
			continue
		}
		if cur == nil {
			cur = &struct {
				path     string
				bare     bool
				detached bool
			}{}
		}
		if strings.HasPrefix(line, "worktree ") {
			cur.path = line[len("worktree "):]
		} else if line == "bare" {
			cur.bare = true
		} else if line == "detached" {
			cur.detached = true
		}
	}
	if cur != nil && cur.path != "" {
		wts = append(wts, *cur)
	}
	return wts
}

func matchesXgoPattern(name string) bool {
	for _, p := range xgoPatterns {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func execGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
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
