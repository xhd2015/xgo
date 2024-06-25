package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/git"
)

// usage:
//
//	go run ./script/git-hooks install
//	go run ./script/git-hooks pre-commit
//	go run ./script/git-hooks pre-commit --no-commit --no-update-version
//	go run ./script/git-hooks post-commit
func main() {
	args := os.Args[1:]
	var cmd string
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	var noCommit bool
	var noUpdateVersion bool
	var amend bool
	for _, arg := range args {
		if arg == "--no-commit" {
			noCommit = true
			continue
		}
		if arg == "--amend" {
			amend = true
			continue
		}
		if arg == "--no-update-version" {
			noUpdateVersion = true
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "unexpected arg: %s\n", arg)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "unrecognized flag: %s\n", arg)
		os.Exit(1)
	}
	if cmd == "" {
		fmt.Fprintf(os.Stderr, "requires command\n")
		os.Exit(1)
	}
	var err error
	if cmd == "install" {
		err = install()
	} else if cmd == "pre-commit" {
		err = preCommitCheck(noCommit, amend, noUpdateVersion)
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

// NOTE: no empty lines in between
const preCommitCmd = `# see: https://stackoverflow.com/questions/19387073/how-to-detect-commit-amend-by-pre-commit-hook
is_amend=$(ps -ocommand= -p $PPID | grep -e '--amend')
# echo "is amend: $is_amend"
# args is always empty
# echo "args: ${args[@]}"
flags=()
if [[ -n $is_amend ]];then
    flags=("${flags[@]}" --amend)
fi
go run ./script/git-hooks pre-commit "${flags[@]}"
`

const postCommitCmdHead = "# xgo check"
const postCommitCmd = "go run ./script/git-hooks post-commit"

func preCommitCheck(noCommit bool, amend bool, noUpdateVersion bool) error {
	gitDir, err := git.ShowTopLevel("")
	if err != nil {
		return err
	}
	rootDir, err := filepath.Abs(gitDir)
	if err != nil {
		return err
	}

	var affectedFiles []string
	const updateRevision = true
	if updateRevision {
		refLast := "HEAD"
		if amend {
			refLast = "HEAD~1"
		}
		commitHash, err := revision.GetCommitHash(rootDir, refLast)
		if err != nil {
			return err
		}

		// due to the nature of git, we cannot
		// know the commit hash of current commit
		// which has not yet happened, so we add
		// suffix "+1" to indicate this
		rev := commitHash + "+1"

		xgoVersionRelFile := revision.GetXgoVersionFile("")
		runtimeVersionRelFile := revision.GetRuntimeVersionFile("")

		xgoVersionFile := filepath.Join(rootDir, xgoVersionRelFile)
		runtimeVersionFile := filepath.Join(rootDir, runtimeVersionRelFile)

		relVersionFiles := []string{xgoVersionRelFile}
		for _, relFile := range relVersionFiles {
			file := filepath.Join(rootDir, relFile)
			content, err := revision.GetFileContent(rootDir, commitHash, relFile)
			if err != nil {
				return err
			}
			version, err := revision.GetVersionNumber(content)
			if err != nil {
				return err
			}
			err = revision.PatchVersionFile(file, "", rev, !noUpdateVersion, version+1)
			if err != nil {
				return err
			}
		}
		err = revision.CopyCoreVersion(xgoVersionFile, runtimeVersionFile)
		if err != nil {
			return err
		}

		affectedFiles = append(affectedFiles, xgoVersionFile, runtimeVersionFile)
	}

	// run generate
	err = cmd.Dir(rootDir).Run("go", "run", "./script/generate", "xgo-runtime")
	if err != nil {
		return err
	}
	affectedFiles = append(affectedFiles, filepath.Join("cmd", "xgo", "runtime_gen"))
	if !noCommit {
		err = cmd.Run("git", append([]string{"add"}, affectedFiles...)...)
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

func install() error {
	// NOTE: is git dir, not toplevel dir when in worktree mode
	gitDir, err := git.GetGitDir("")
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	err = os.MkdirAll(hooksDir, 0755)
	if err != nil {
		return err
	}

	err = installHook(filepath.Join(hooksDir, "pre-commit"), preCommitCmdHead, preCommitCmd)
	if err != nil {
		return fmt.Errorf("pre-commit: %w", err)
	}

	err = installHook(filepath.Join(hooksDir, "post-commit"), postCommitCmdHead, postCommitCmd)
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
	if needChmod {
		err := os.Chmod(hookFile, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
