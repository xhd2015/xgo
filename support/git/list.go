// functions copied from https://github.com/xhd2015/gitops
package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

const COMMIT_WORKING = "WORKING"

// ListFileUpdates lists all updated files in the given directory
func ListFileUpdates(dir string, ref string, compareRef string, patterns []string) ([]string, error) {
	content, err := diffFiles(dir, ref, compareRef, []string{"--diff-filter=ACDMR"}, patterns)
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(content), nil
}

// git diff --diff-filter=M --name-only --ignore-submodules "$compareRef" "$ref" -- ${patterns} || true
func ListModifiedFiles(dir string, ref string, compareRef string, patterns []string) ([]string, error) {
	content, err := diffFiles(dir, ref, compareRef, []string{"--diff-filter=M"}, patterns)
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(content), nil
}

// doc: https://git-scm.com/docs/git-diff
// extraFlags: --diff-filter=A
// --diff-filter=[(A|C|D|M|R|T|U|X|B)...[*]]
// Select only files that are Added (A), Copied (C), Deleted (D), Modified (M), Renamed (R), have their type (i.e. regular file, symlink, submodule, …​) changed (T), are Unmerged (U), are Unknown (X), or have had their pairing Broken (B). Any combination of the filter characters (including none) can be used. When * (All-or-none) is added to the combination, all paths are selected if there is any file that matches other criteria in the comparison; if there is no file that matches other criteria, nothing is selected.
// Also, these upper-case letters can be downcased to exclude. E.g. --diff-filter=ad excludes added and deleted paths.
// Note that not all diffs can feature all types. For instance, copied and renamed entries cannot appear if detection for those types is disabled.
func diffFiles(dir string, ref string, compareRef string, extraFlags []string, patterns []string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("requires ref")
	}
	if compareRef == "" {
		return "", fmt.Errorf("requires compareRef")
	}

	// git diff --diff-filter=A --name-only --ignore-submodules "$compareRef" "$ref" -- ${patterns} || true
	// --relative: always respect current dir. without this, when dir is a sub dir from toplevle, diff still returns full path
	args := []string{"-c", "core.fileMode=false", "diff", "--relative"}
	args = append(args, extraFlags...)
	args = append(args, "--name-only", "--ignore-submodules", compareRef)
	if ref != COMMIT_WORKING {
		args = append(args, ref)
	}

	if len(patterns) > 0 {
		args = append(args, "--")
		args = append(args, patterns...)
	}

	output, err := cmd.Dir(dir).Output("git", args...)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", nil
		}
		return "", err
	}
	return output, nil
}

func splitLinesFilterEmpty(s string) []string {
	list := strings.Split(s, "\n")
	idx := 0
	for _, e := range list {
		e = strings.TrimSpace(e)
		if e != "" {
			list[idx] = e
			idx++
		}
	}
	return list[:idx]
}
