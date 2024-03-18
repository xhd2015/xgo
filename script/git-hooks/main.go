package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/strutil"
)

// usage:
//
//	go run ./script/git-hooks install
//	go run ./script/git-hooks pre-commit
//	go run ./script/git-hooks pre-commit --no-commit
//	go run ./script/git-hooks post-commit
func main() {
	args := os.Args[1:]
	var cmd string
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	var noCommit bool
	for _, arg := range args {
		if arg == "--no-commit" {
			noCommit = true
			continue
		}
	}
	if cmd == "" {
		fmt.Fprintf(os.Stderr, "requires command\n")
		os.Exit(1)
	}
	var err error
	if cmd == "install" {
		err = install()
	} else if cmd == "pre-commit" {
		err = preCommitCheck(noCommit)
	} else if cmd == "post-commit" {
		err = postCommitCheck(noCommit)
	} else {
		fmt.Fprintf(os.Stderr, "unrecognized command: %s\n", cmd)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

const preCommitCmdHead = "# xgo check"
const preCommitCmd = "go run ./script/git-hooks pre-commit"

const postCommitCmdHead = "# xgo check"
const postCommitCmd = "go run ./script/git-hooks post-commit"

func getGitDir() (string, error) {
	return cmd.Output("git", "rev-parse", "--git-dir")
}

func preCommitCheck(noCommit bool) error {
	gitDir, err := getGitDir()
	if err != nil {
		return err
	}
	rootDir := filepath.Dir(gitDir)
	if rootDir == "" {
		return fmt.Errorf("invalid git dir:%s", gitDir)
	}
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		return err
	}

	commitHash, err := cmd.Output("git", "log", "--format=%H", "-1", "HEAD")
	if err != nil {
		return err
	}
	// due to the nature of git, we cannot
	// know the commit hash of current commit
	// which has not yet happened, so we add
	// suffix "+1" to indicate this
	revision := commitHash + "+1"
	xgoVersionFile := filepath.Join(rootDir, "cmd", "xgo", "version.go")
	runtimeVersionFile := filepath.Join(rootDir, "runtime", "core", "version.go")

	err = fileutil.Patch(xgoVersionFile, func(data []byte) ([]byte, error) {
		content := string(data)
		newContent, err := replaceRevision(content, revision)
		if err != nil {
			return nil, err
		}
		return []byte(newContent), nil
	})
	if err != nil {
		return fmt.Errorf("cmd/xgo/version.go: %w", err)
	}

	err = fileutil.Patch(runtimeVersionFile, func(data []byte) ([]byte, error) {
		content := string(data)
		newContent, err := replaceRevision(content, revision)
		if err != nil {
			return nil, err
		}
		return []byte(newContent), nil
	})
	if err != nil {
		return fmt.Errorf("runtime/core/version.go: %w", err)
	}

	if !noCommit {
		err = cmd.Run("git", "add", xgoVersionFile, runtimeVersionFile)
		if err != nil {
			return nil
		}

		// --no-verify: skip pre-commit and post-commit checks
		// err = cmd.Run("git", "commit", "--no-verify", "--amend", "--no-edit")
		// if err != nil {
		// 	return nil
		// }
	}

	return nil
}

func postCommitCheck(noCommit bool) error {
	// do nothing
	return nil
}

func replaceRevision(s string, revision string) (string, error) {
	if strings.Contains(revision, `"`) {
		return "", fmt.Errorf("revision connot have \": %s", revision)
	}
	lines := strings.Split(s, "\n")
	n := len(lines)
	lineIdx := -1
	byteIdx := -1
	for i := 0; i < n; i++ {
		line := lines[i]
		idx := strutil.IndexSequence(line, []string{"const", "REVISION", "="})
		if idx >= 0 {
			lineIdx = i
			byteIdx = idx
			break
		}
	}
	if lineIdx < 0 {
		return "", fmt.Errorf("variable REVISION not found")
	}
	line := lines[lineIdx]
	qIdx := strings.Index(line[byteIdx+1:], `"`)
	if qIdx < 0 {
		return "", fmt.Errorf("invalid REVISION variable, missing \"")
	}
	qIdx += byteIdx + 1
	endIdx := strings.Index(line[qIdx+1:], `"`)
	if endIdx < 0 {
		return "", fmt.Errorf("invalid REVISION variable, missing ending \"")
	}
	endIdx += qIdx + 1
	line = line[:qIdx+1] + revision + line[endIdx:]

	lines[lineIdx] = line

	return strings.Join(lines, "\n"), nil
}

func install() error {
	gitDir, err := getGitDir()
	if err != nil {
		return err
	}

	err = installHook(filepath.Join(gitDir, "hooks", "pre-commit"), preCommitCmdHead, preCommitCmd)
	if err != nil {
		return fmt.Errorf("pre-commit: %w", err)
	}

	err = installHook(filepath.Join(gitDir, "hooks", "post-commit"), postCommitCmdHead, postCommitCmd)
	if err != nil {
		return fmt.Errorf("post-commit: %w", err)
	}
	return nil
}

func installHook(hookFile string, head string, cmd string) error {
	var needChmod bool
	err := fileutil.Patch(hookFile, func(data []byte) ([]byte, error) {
		if len(data) == 0 {
			needChmod = true
		}
		content := string(data)
		lines := strings.Split(content, "\n")
		idx := -1
		n := len(lines)
		for i := 0; i < n; i++ {
			if strings.Contains(lines[i], head) {
				idx = i
				break
			}
		}
		if idx < 0 {
			// insert
			lines = append(lines, head, cmd, "")
		} else {
			// replace
			endIdx := idx + 1
			for ; endIdx < n; endIdx++ {
				if strings.TrimSpace(lines[endIdx]) == "" {
					break
				}
			}
			oldLines := lines
			lines = lines[:idx]
			lines = append(lines, head, cmd, "")
			if endIdx < n {
				lines = append(lines, oldLines[endIdx:]...)
			}
		}

		return []byte(strings.Join(lines, "\n")), nil
	})

	if err != nil {
		return err
	}

	// chmod to what? it is 0755 already
	if needChmod && false {
		err := os.Chmod(hookFile, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
