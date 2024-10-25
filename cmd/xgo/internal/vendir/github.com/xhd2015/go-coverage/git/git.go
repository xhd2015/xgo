package git

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"sync"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/sh"
)

// a special commit that referes to current working directory(a pseduo commit)
const COMMIT_WORKING = "WORKING"

// find rename
// git diff --find-renames --diff-filter=R   HEAD~10 HEAD|grep -A 3 '^diff --git a/'|grep rename
// FindRenames returns a mapping from new name to old name
func FindRenames(dir string, oldCommit string, newCommit string) (map[string]string, error) {
	repo := &GitRepo{Dir: dir}
	return repo.FindRenames(oldCommit, newCommit)
}
func FindRenamesV2(dir string, oldCommit string, newCommit string, fn func(newFile string, oldFile string, percent string)) error {
	repo := &GitRepo{Dir: dir}
	return repo.FindRenamesV2(oldCommit, newCommit, fn)
}

func FindUpdate(dir string, oldCommit string, newCommit string) ([]string, error) {
	repo := &GitRepo{Dir: dir}
	return repo.FindUpdate(oldCommit, newCommit)
}

func FindUpdateAndRenames(dir string, oldCommit string, newCommit string) (newToOld map[string]string, err error) {
	repo := &GitRepo{Dir: dir}
	updates, err := repo.FindUpdate(oldCommit, newCommit)
	if err != nil {
		return nil, err
	}
	m, err := repo.FindRenames(oldCommit, newCommit)
	if err != nil {
		return nil, err
	}
	for _, u := range updates {
		if _, ok := m[u]; ok {
			return nil, fmt.Errorf("invalid file: %s found both renamed and updated", u)
		}
		m[u] = u
	}
	return m, nil
}
func FindAll(dir string, oldCommit string, newCommit string) (allFiles []string, newToOld map[string]string, err error) {
	newToOld, err = FindUpdateAndRenames(dir, oldCommit, newCommit)
	if err != nil {
		return
	}
	allFiles, err = NewSnapshot(dir, newCommit).ListFiles()
	return
}

type GitRepo struct {
	Dir string
}

func NewGitRepo(dir string) *GitRepo {
	return &GitRepo{
		Dir: dir,
	}
}
func NewSnapshot(dir string, commit string) *GitSnapshot {
	return &GitSnapshot{
		Dir:    dir,
		Commit: commit,
	}
}

func QuoteCommit(commit string) string {
	if commit == COMMIT_WORKING {
		return ""
	}
	return sh.Quote(getRef(commit))
}

func (c *GitRepo) FindUpdate(oldCommit string, newCommit string) ([]string, error) {
	cmd := fmt.Sprintf(`git -C %s diff --diff-filter=M --name-only --ignore-submodules %s %s`, sh.Quote(c.Dir), QuoteCommit(oldCommit), QuoteCommit(newCommit))
	stdout, _, err := sh.RunBashCmdOpts(cmd, sh.RunBashOptions{
		NeedStdOut: true,
	})
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(stdout), nil
}
func (c *GitRepo) FindAdded(oldCommit string, newCommit string) ([]string, error) {
	cmd := fmt.Sprintf(`git -C %s diff --diff-filter=A --name-only --ignore-submodules %s %s`, sh.Quote(c.Dir), QuoteCommit(oldCommit), QuoteCommit(newCommit))
	stdout, _, err := sh.RunBashCmdOpts(cmd, sh.RunBashOptions{
		NeedStdOut: true,
	})
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(stdout), nil
}
func (c *GitRepo) FindRenames(oldCommit string, newCommit string) (map[string]string, error) {
	// example:
	// 	$ git diff --find-renames --diff-filter=R   HEAD~10 HEAD
	// diff --git a/test/stubv2/boot/boot.go b/test/stub/boot/boot.go
	// similarity index 94%
	// rename from test/stubv2/boot/boot.go
	// rename to test/stub/boot/boot.go
	// index e0e86051..56c49801 100644
	// --- a/test/stubv2/boot/boot.go
	// +++ b/test/stub/boot/boot.go
	// @@ -4,8 +4,10 @@ import (
	cmd := fmt.Sprintf(`git -C %s diff --find-renames --diff-filter=R --ignore-submodules %s %s|grep -A 3 '^diff --git a/'|grep -E '^rename' || true`, sh.Quote(c.Dir), QuoteCommit(oldCommit), QuoteCommit(newCommit))
	stdout, stderr, err := sh.RunBashCmdOpts(cmd, sh.RunBashOptions{
		// Verbose:    true,
		NeedStdOut: true,
		// NeedStdErr: true,
	})
	// fmt.Printf("stderr:%v", stderr)
	_ = stderr
	if err != nil {
		return nil, err
	}
	lines := splitLinesFilterEmpty(stdout)
	if len(lines)%2 != 0 {
		return nil, fmt.Errorf("internal error, expect git return rename pairs, found:%d", len(lines))
	}

	m := make(map[string]string, len(lines)/2)
	for i := 0; i < len(lines); i += 2 {
		from := strings.TrimPrefix(lines[i], "rename from ")
		to := strings.TrimPrefix(lines[i+1], "rename to ")

		m[to] = from
	}
	return m, nil
}

// without --ignore-submodules, git diff may include diff of dirs
func (c *GitRepo) FindRenamesV2(oldCommit string, newCommit string, fn func(newFile string, oldFile string, percent string)) error {
	cmd := fmt.Sprintf(`git -C %s diff --diff-filter=R --summary --ignore-submodules %s %s || true`, sh.Quote(c.Dir), QuoteCommit(oldCommit), QuoteCommit(newCommit))
	stdout, stderr, err := sh.RunBashCmdOpts(cmd, sh.RunBashOptions{
		// Verbose:    true,
		NeedStdOut: true,
		// NeedStdErr: true,
	})
	// fmt.Printf("stderr:%v", stderr)
	_ = stderr
	if err != nil {
		return err
	}
	parseRenames(stdout, fn)
	return nil
}

func parseRenames(renamedFilesWithSummary string, fn func(newFile string, oldFile string, percent string)) {
	renames := splitLinesFilterEmpty(renamedFilesWithSummary)
	for _, line := range renames {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "rename ") {
			continue
		}
		s := strings.TrimSpace(line[len("rename "):])

		// parse percent
		//  rename src/{module_funder/id/submodule_funder_seabank/cl => module_funder_bke/id}/bcl/task/repair_clawback_task.go (64%)
		var percent string
		pEidx := strings.LastIndex(s, ")")
		if pEidx >= 0 {
			s = strings.TrimSpace(s[:pEidx])
			pIdx := strings.LastIndex(s, "(")
			if pIdx >= 0 {
				percent = strings.TrimSpace(s[pIdx+1:])
				s = strings.TrimSpace(s[:pIdx])
			}
		}

		bIdx := strings.Index(s, "{")
		if bIdx < 0 {
			continue
		}
		bEIdx := strings.LastIndex(s, "}")
		if bEIdx < 0 {
			continue
		}
		prefix := s[:bIdx]
		var suffix string
		if bEIdx+1 < len(s) {
			suffix = s[bEIdx+1:]
		}
		s = s[bIdx+1 : bEIdx]
		sep := " => "
		toIdx := strings.Index(s, sep)
		if toIdx < 0 {
			continue
		}
		oldPath := s[:toIdx]
		var newPath string
		if toIdx+len(sep) < len(s) {
			newPath = s[toIdx+len(sep):]
		}

		file := joinPath(prefix, newPath, suffix)
		oldFile := joinPath(prefix, oldPath, suffix)

		fn(file, oldFile, percent)
	}
}

func joinPath(p ...string) string {
	j := 0
	for i := 0; i < len(p); i++ {
		e := strings.TrimPrefix(p[i], "/")
		e = strings.TrimSuffix(e, "/")
		if e != "" {
			p[j] = e
			j++
		}
	}
	return strings.Join(p[:j], "/")
}

type GitSnapshot struct {
	Dir    string
	Commit string

	filesInit sync.Once
	files     []string
	filesErr  error
	fileMap   map[string]bool
}

func (c *GitSnapshot) GetContent(file string) (string, error) {
	normFile := strings.TrimPrefix(file, "./")
	if normFile == "" {
		return "", fmt.Errorf("invalid file:%v", file)
	}

	c.ensureInitList()
	if !c.fileMap[normFile] {
		return "", fmt.Errorf("not a file, maybe a dir:%v", file)
	}
	if c.Commit == COMMIT_WORKING {
		content, err := ioutil.ReadFile(path.Join(c.Dir, normFile))
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	content, _, err := sh.RunBashWithOpts([]string{
		fmt.Sprintf("git -C %s cat-file -p %s:%s", sh.Quote(c.Dir), sh.Quote(c.ref()), sh.Quote(normFile)),
	}, sh.RunBashOptions{
		NeedStdOut: true,
	})
	return content, err
}
func (c *GitSnapshot) ListFiles() ([]string, error) {
	c.ensureInitList()
	return c.files, c.filesErr
}
func (c *GitSnapshot) ensureInitList() {
	withTree := ""
	if c.Commit != COMMIT_WORKING {
		withTree = fmt.Sprintf("--with-tree %s", QuoteCommit(c.Commit))
	}
	c.filesInit.Do(func() {
		stdout, _, err := sh.RunBashWithOpts([]string{
			fmt.Sprintf("git -C %s ls-files %s", sh.Quote(c.Dir), withTree),
		}, sh.RunBashOptions{
			Verbose:    true,
			NeedStdOut: true,
		})
		if err != nil {
			c.filesErr = err
			return
		}
		c.files = splitLinesFilterEmpty(stdout)
		c.fileMap = make(map[string]bool)
		for _, e := range c.files {
			c.fileMap[e] = true
		}
	})
}

func (c *GitSnapshot) ref() string {
	return getRef(c.Commit)
}

func getRef(commit string) string {
	if commit == "" {
		return "HEAD"
	}
	return commit
}

func splitLinesFilterEmpty(s string) []string {
	list := strings.Split(s, "\n")
	idx := 0
	for _, e := range list {
		if e != "" {
			list[idx] = e
			idx++
		}
	}
	return list[:idx]
}
