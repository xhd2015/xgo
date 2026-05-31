package gitgoroot

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/goinfo"
)

const (
	downloadListURL   = "https://go.dev/dl"
	downloadURLPrefix = "https://go.dev/dl/"
	incompleteMarker  = ".xgo-incomplete"
)

func ResolveVersion(version string) (string, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return "", fmt.Errorf("empty Go version")
	}
	version = strings.TrimPrefix(version, "go")

	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid Go version: %q", version)
	}
	if len(parts) == 3 {
		if !goinfo.IsValidVersion(version) {
			return "", fmt.Errorf("invalid Go version: %q", version)
		}
		return "go" + version, nil
	}

	allVersions, err := fetchVersionList()
	if err != nil {
		return "", fmt.Errorf("fetch version list: %w", err)
	}

	prefix := version + "."
	var latest string
	for _, v := range allVersions {
		if strings.HasPrefix(v, prefix) {
			if latest == "" || goinfo.CompareVersion(v[len(prefix):], latest[len(prefix):]) > 0 {
				latest = v
			}
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no patch version found for %q", version)
	}
	return "go" + latest, nil
}

func fetchVersionList() ([]string, error) {
	resp, err := http.Get(downloadListURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return parseDownloadVersions(string(body)), nil
}

func parseDownloadVersions(htmlContent string) []string {
	var goVersions []string
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

func EnsureGitGoroot(baseDir, version string, bootstrapGoroot string) (string, error) {
	versionName, err := ResolveVersion(version)
	if err != nil {
		return "", err
	}

	goroot := filepath.Join(baseDir, versionName)
	if _, err := os.Stat(goroot); err == nil {
		if _, err := os.Stat(filepath.Join(goroot, incompleteMarker)); err == nil {
			if err := os.RemoveAll(goroot); err != nil {
				return "", fmt.Errorf("remove incomplete goroot %s: %w", goroot, err)
			}
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("stat incomplete marker: %w", err)
		}
	}

	if _, err := os.Stat(goroot); err == nil {
		valid, err := IsGitGorootValid(goroot, versionName)
		if err != nil {
			return "", err
		}
		if !valid {
			return "", fmt.Errorf("existing goroot %s is not in expected state", goroot)
		}
		return goroot, nil
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath.Join(goroot, incompleteMarker), nil, 0644); err != nil {
		if err := os.MkdirAll(goroot, 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(goroot, incompleteMarker), nil, 0644); err != nil {
			return "", err
		}
	}

	cleanup := func() { os.RemoveAll(goroot) }

	if err := initGitGoroot(goroot, versionName, bootstrapGoroot); err != nil {
		cleanup()
		return "", err
	}

	if err := os.Remove(filepath.Join(goroot, incompleteMarker)); err != nil {
		return "", fmt.Errorf("remove incomplete marker: %w", err)
	}

	return goroot, nil
}

func initGitGoroot(goroot, versionName, bootstrapGoroot string) error {
	if err := downloadAndExtract(goroot, versionName); err != nil {
		return err
	}

	if err := runGit(goroot, "init", "-q"); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(goroot, ".gitignore"), []byte(".DS_Store\n"), 0644); err != nil {
		return err
	}

	if err := buildGoToolchain(goroot, bootstrapGoroot); err != nil {
		return err
	}

	if err := runGit(goroot, "add", "."); err != nil {
		return err
	}

	commitMsg := "init " + versionName
	if err := runGit(goroot,
		"-c", "user.name=xgo",
		"-c", "user.email=xgo@example.invalid",
		"commit", "-q", "-m", commitMsg,
	); err != nil {
		return err
	}

	if err := runGit(goroot, "branch", "-m", versionName); err != nil {
		return err
	}

	// Ignore file mode (executable bit) diffs for all worktrees.
	// The initial commit already captured correct modes.
	if err := runGit(goroot, "config", "core.filemode", "false"); err != nil {
		return err
	}

	return nil
}

func runGit(dir string, args ...string) error {
	allArgs := append([]string{"-c", "core.hooksPath="}, args...)
	return runLogged(dir, nil, "git", allArgs...)
}

func runGitOutput(dir string, args ...string) (string, error) {
	allArgs := append([]string{"-c", "core.hooksPath="}, args...)
	return runOutput(dir, "git", allArgs...)
}

func downloadAndExtract(goroot, versionName string) error {
	tarParent := filepath.Dir(goroot)
	tarName := versionName + ".src.tar.gz"
	tarPath := filepath.Join(tarParent, tarName)
	tarURL := downloadURLPrefix + tarName

	if err := runLogged(tarParent, nil, "curl", "-fL", "-o", tarPath, tarURL); err != nil {
		return fmt.Errorf("download %s: %w", tarURL, err)
	}
	defer os.Remove(tarPath)

	tmpExtract := filepath.Join(tarParent, versionName+".extract-tmp")
	if err := os.MkdirAll(tmpExtract, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(tmpExtract)

	if err := runLogged(tarParent, nil, "tar", "-xzf", tarPath, "-C", tmpExtract); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	extractedGo := filepath.Join(tmpExtract, "go")
	if _, err := os.Stat(extractedGo); err != nil {
		return fmt.Errorf("extracted dir not found: %w", err)
	}

	files, err := os.ReadDir(extractedGo)
	if err != nil {
		return fmt.Errorf("read extracted dir: %w", err)
	}
	for _, f := range files {
		src := filepath.Join(extractedGo, f.Name())
		dst := filepath.Join(goroot, f.Name())
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("move %s to goroot: %w", f.Name(), err)
		}
	}

	return nil
}

func buildGoToolchain(goroot, bootstrapGoroot string) error {
	if bootstrapGoroot == "" {
		out, err := runOutput("", "go", "env", "GOROOT")
		if err != nil {
			return fmt.Errorf("get bootstrap GOROOT: %w (set GOROOT_BOOTSTRAP env)", err)
		}
		bootstrapGoroot = out
	}

	return runLogged(filepath.Join(goroot, "src"), []string{"GOROOT_BOOTSTRAP=" + bootstrapGoroot}, "./make.bash")
}

func IsGitGorootValid(gitGoroot, versionName string) (bool, error) {
	if _, err := os.Stat(filepath.Join(gitGoroot, incompleteMarker)); err == nil {
		return false, nil
	}

	branch, err := runGitOutput(gitGoroot, "branch", "--show-current")
	if err != nil {
		return false, fmt.Errorf("get branch: %w", err)
	}
	if branch != versionName {
		return false, nil
	}

	count, err := runGitOutput(gitGoroot, "rev-list", "--count", "HEAD")
	if err != nil {
		return false, fmt.Errorf("count commits: %w", err)
	}
	if strings.TrimSpace(count) != "1" {
		return false, nil
	}

	commitMsg, err := runGitOutput(gitGoroot, "log", "-1", "--format=%s")
	if err != nil {
		return false, fmt.Errorf("get commit message: %w", err)
	}
	expectedMsg := "init " + versionName
	if commitMsg != expectedMsg {
		return false, nil
	}

	return true, nil
}

func SpawnWorktree(gitGoroot string) (string, func(), error) {
	versionName := filepath.Base(gitGoroot)

	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, fmt.Errorf("generate random: %w", err)
	}
	randomSuffix := hex.EncodeToString(randomBytes)

	worktreePath := filepath.Join(os.TempDir(), "xgo-wt-"+versionName+"-"+randomSuffix)

	if err := runGit("", "-C", gitGoroot, "worktree", "add", "--detach", worktreePath, "HEAD"); err != nil {
		return "", nil, fmt.Errorf("create worktree: %w", err)
	}

	cleanup := func() {
		if err := runGit("", "-C", gitGoroot, "worktree", "remove", "--force", worktreePath); err != nil {
			fmt.Fprintf(os.Stderr, "cleanup: remove worktree: %v\n", err)
		}
		if err := os.RemoveAll(worktreePath); err != nil {
			fmt.Fprintf(os.Stderr, "cleanup: remove worktree dir: %v\n", err)
		}
	}

	return worktreePath, cleanup, nil
}

func runOutput(dir string, cmdName string, args ...string) (string, error) {
	return cmd.New().Dir(dir).Output(cmdName, args...)
}

func runLogged(dir string, env []string, cmdName string, args ...string) error {
	logf("+ %s", shellQuoteArgs(cmdName, args))
	return cmd.New().Dir(dir).Env(env).Debug().Run(cmdName, args...)
}

func logf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[gitgoroot] "+format+"\n", args...)
}

func shellQuoteArgs(cmdName string, args []string) string {
	parts := append([]string{cmdName}, args...)
	for i, p := range parts {
		parts[i] = shellQuote(p)
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		if !(r == '/' || r == '.' || r == '-' || r == '_' || r == '=' || r == ':' ||
			('0' <= r && r <= '9') ||
			('A' <= r && r <= 'Z') ||
			('a' <= r && r <= 'z')) {
			return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		}
	}
	return s
}

