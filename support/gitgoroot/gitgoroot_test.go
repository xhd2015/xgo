package gitgoroot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveVersion_Exact(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"go1.24.2", "go1.24.2"},
		{"1.24.2", "go1.24.2"},
		{"go1.22.0", "go1.22.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ResolveVersion(tt.input)
			if err != nil {
				t.Fatalf("ResolveVersion(%q) error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Fatalf("ResolveVersion(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveVersion_Minor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires network in short mode")
	}

	got, err := ResolveVersion("go1.24")
	if err != nil {
		t.Fatalf("ResolveVersion(\"go1.24\") error: %v", err)
	}
	if !strings.HasPrefix(got, "go1.24.") {
		t.Fatalf("ResolveVersion(\"go1.24\") = %q, want prefix go1.24.", got)
	}

	parts := strings.Split(got, ".")
	if len(parts) != 3 {
		t.Fatalf("ResolveVersion(\"go1.24\") = %q, want three components", got)
	}
}

func TestResolveVersion_Invalid(t *testing.T) {
	tests := []string{
		"",
		"go",
		"1",
		"go1",
		"go1.24.2.3",
		"notaversion",
	}

	for _, input := range tests {
		t.Run("invalid_"+input, func(t *testing.T) {
			_, err := ResolveVersion(input)
			if err == nil {
				t.Fatalf("ResolveVersion(%q) should have returned error", input)
			}
		})
	}
}

func TestResolveVersion_Nonexistent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires network in short mode")
	}

	_, err := ResolveVersion("go1.99")
	if err == nil {
		t.Fatal("ResolveVersion(\"go1.99\") should have returned error")
	}
}

func TestIsGitGorootValid_Valid(t *testing.T) {
	dir := t.TempDir()
	initTestGitRepo(t, dir, "go1.24.2")

	valid, err := IsGitGorootValid(dir, "go1.24.2")
	if err != nil {
		t.Fatalf("IsGitGorootValid error: %v", err)
	}
	if !valid {
		t.Fatal("expected valid git goroot")
	}
}

func TestIsGitGorootValid_WrongBranch(t *testing.T) {
	dir := t.TempDir()
	initTestGitRepo(t, dir, "go1.24.2")
	exec.Command("git", "-C", dir, "branch", "-m", "wrong-branch").Run()

	valid, err := IsGitGorootValid(dir, "go1.24.2")
	if err != nil {
		t.Fatalf("IsGitGorootValid error: %v", err)
	}
	if valid {
		t.Fatal("expected invalid git goroot due to wrong branch")
	}
}

func TestIsGitGorootValid_ExtraCommit(t *testing.T) {
	dir := t.TempDir()
	initTestGitRepo(t, dir, "go1.24.2")
	os.WriteFile(filepath.Join(dir, "extra.txt"), []byte("extra\n"), 0644)
	exec.Command("git", "-C", dir, "add", "extra.txt").Run()
	exec.Command("git", "-C", dir, "-c", "user.name=test", "-c", "user.email=test@test", "commit", "-q", "-m", "extra commit").Run()

	valid, err := IsGitGorootValid(dir, "go1.24.2")
	if err != nil {
		t.Fatalf("IsGitGorootValid error: %v", err)
	}
	if valid {
		t.Fatal("expected invalid git goroot due to extra commit")
	}
}

func TestIsGitGorootValid_WrongCommitMessage(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(dir, 0755)
	gitRunHelper(t, dir, "git", "init", "-q")
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test\n"), 0644)
	gitRunHelper(t, dir, "git", "add", "test.txt")
	gitRunHelper(t, dir, "git", "-c", "user.name=test", "-c", "user.email=test@test", "commit", "-q", "-m", "different message")
	gitRunHelper(t, dir, "git", "branch", "-m", "go1.24.2")

	valid, err := IsGitGorootValid(dir, "go1.24.2")
	if err != nil {
		t.Fatalf("IsGitGorootValid error: %v", err)
	}
	if valid {
		t.Fatal("expected invalid git goroot due to wrong commit message")
	}
}

func TestIsGitGorootValid_IncompleteMarker(t *testing.T) {
	dir := t.TempDir()
	initTestGitRepo(t, dir, "go1.24.2")
	os.WriteFile(filepath.Join(dir, incompleteMarker), nil, 0644)

	valid, err := IsGitGorootValid(dir, "go1.24.2")
	if err != nil {
		t.Fatalf("IsGitGorootValid error: %v", err)
	}
	if valid {
		t.Fatal("expected invalid git goroot due to incomplete marker")
	}
}

func TestSpawnWorktree(t *testing.T) {
	dir := t.TempDir()
	initTestGitRepo(t, dir, "go1.24.2")

	worktree, cleanup, err := SpawnWorktree(dir)
	if err != nil {
		t.Fatalf("SpawnWorktree error: %v", err)
	}
	defer cleanup()

	if _, err := os.Stat(worktree); err != nil {
		t.Fatalf("worktree dir does not exist: %v", err)
	}

	if _, err := os.Stat(filepath.Join(worktree, "test.txt")); err != nil {
		t.Fatalf("test.txt not found in worktree: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(worktree, "test.txt"))
	if err != nil {
		t.Fatalf("read test.txt: %v", err)
	}
	if string(data) != "test\n" {
		t.Fatalf("test.txt content = %q, want %q", string(data), "test\n")
	}
}

func TestSpawnWorktree_Isolation(t *testing.T) {
	dir := t.TempDir()
	initTestGitRepo(t, dir, "go1.24.2")

	wt1, cleanup1, err := SpawnWorktree(dir)
	if err != nil {
		t.Fatalf("SpawnWorktree 1 error: %v", err)
	}
	defer cleanup1()

	wt2, cleanup2, err := SpawnWorktree(dir)
	if err != nil {
		t.Fatalf("SpawnWorktree 2 error: %v", err)
	}
	defer cleanup2()

	if wt1 == wt2 {
		t.Fatal("expected different worktree paths")
	}

	os.WriteFile(filepath.Join(wt1, "modified.txt"), []byte("from wt1\n"), 0644)

	if _, err := os.Stat(filepath.Join(wt2, "modified.txt")); err == nil {
		t.Fatal("modification in worktree 1 leaked to worktree 2")
	}
}

func TestSpawnWorktree_Cleanup(t *testing.T) {
	dir := t.TempDir()
	initTestGitRepo(t, dir, "go1.24.2")

	worktree, cleanup, err := SpawnWorktree(dir)
	if err != nil {
		t.Fatalf("SpawnWorktree error: %v", err)
	}

	cleanup()

	if _, err := os.Stat(worktree); !os.IsNotExist(err) {
		t.Fatal("worktree dir should be removed after cleanup")
	}
}

func initTestGitRepo(t *testing.T, dir, versionName string) {
	t.Helper()
	os.MkdirAll(dir, 0755)
	gitRunHelper(t, dir, "git", "init", "-q")
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test\n"), 0644)
	gitRunHelper(t, dir, "git", "add", "test.txt")
	gitRunHelper(t, dir,
		"git", "-c", "user.name=test", "-c", "user.email=test@test",
		"commit", "-q", "-m", "init "+versionName,
	)
	gitRunHelper(t, dir, "git", "branch", "-m", versionName)
}

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
}

func gitRunHelper(t *testing.T, dir string, args ...string) {
	t.Helper()
	skipIfNoGit(t)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args[1:], err, string(out))
	}
}
