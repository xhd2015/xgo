package gitops

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/model"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/sh"
)

const COMMIT_WORKING = "WORKING"

type FileDetail = model.FileDetail

type DiffCommitOptions struct {
	PathPatterns []string `json:"pathPatterns"` // only include these patterns, example: src/**/*.go
}

// DiffCommit finds file changes between two files
func DiffCommit(dir string, ref string, compareRef string, options *DiffCommitOptions) (fileDetailsMap map[string]*FileDetail, err error) {
	if ref == "" {
		err = fmt.Errorf("requires ref")
		return
	}
	if compareRef == "" {
		err = fmt.Errorf("requires compareRef")
		return
	}
	if compareRef == COMMIT_WORKING {
		err = fmt.Errorf("compareRef cannot be WORKING commit")
		return
	}

	var patternArg string
	if options != nil && len(options.PathPatterns) > 0 {
		patternArg = " -- " + sh.Quotes(options.PathPatterns...)
	}

	// all files
	// git ls-files --with-tree new

	// new
	// git diff --diff-filter=A --name-only old new -- 'src/**/*.go'

	// update
	// git diff --diff-filter=M --name-only old new

	// rename
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

	// command: list renames, and grep rename after each diff
	// git diff --find-renames --diff-filter=R %s %s|grep -A 3 '^diff --git a/'|grep -E '^rename' || true

	defRef := "ref="

	var collectUntrackedFilesCmd string

	var withTreeRef string
	var workingFlags string
	if ref != COMMIT_WORKING {
		defRef = fmt.Sprintf("ref=$(git rev-parse --verify --quiet %s || true)", sh.Quote(ref))
		withTreeRef = "--with-tree $ref"
	} else {
		// -c cached (default)
		// -o others,including untracked
		workingFlags = "--exclude-standard -co"

		collectUntrackedFilesCmd = "git ls-files --no-empty-directory --exclude-standard --others  --full-name" + patternArg
	}
	res, err := RunCommand(dir, func(commands []string) []string {
		return append(commands, []string{
			defRef,
			fmt.Sprintf("compareRef=$(git rev-parse --verify --quiet %s || true)", sh.Quote(compareRef)),
			// all new files
			"echo 'all-new-files:'",
			fmt.Sprintf(`git ls-files %s %s %s`, withTreeRef, workingFlags, patternArg),
			"echo -ne '\r\r\r\r\r\r'",

			// all old files
			"echo 'all-old-files:'",
			fmt.Sprintf(`git ls-files --with-tree "$compareRef" %s`, patternArg),
			"echo -ne '\r\r\r\r\r\r'",

			// added files
			"echo 'added-files:'",
			fmt.Sprintf(`git diff --diff-filter=A --name-only --ignore-submodules "$compareRef" $ref -- %s|| true`, patternArg),
			"echo -ne '\r\r\r\r\r\r'",

			// untracked files
			"echo 'untracked-files:'",
			collectUntrackedFilesCmd,
			"echo -ne '\r\r\r\r\r\r'",

			// modified files
			// add -c core.fileMode=false to ignore mode change
			"echo 'modified-files:'",
			fmt.Sprintf(`git -c core.fileMode=false diff --diff-filter=M --name-only --ignore-submodules "$compareRef" $ref -- %s|| true`, patternArg),
			"echo -ne '\r\r\r\r\r\r'",

			// // renamed files
			// "echo 'renamed-files:'",
			// `git diff --find-renames --diff-filter=R "$compareRef" "$ref"|grep -A 3 '^diff --git a/'|grep -E '^rename' || true`,

			// renamed files with summary
			"echo 'renamed-files-with-summary:'",
			//  rename src/{module_funder/id/submodule_funder_seabank/cl => module_funder_bke/id}/bcl/task/repair_clawback_task.go (64%)
			fmt.Sprintf(`git diff --diff-filter=R --summary --ignore-submodules "$compareRef" $ref -- %s|| true`, patternArg),
		}...)
	})
	if err != nil {
		return nil, err
	}

	var allNewFiles []string
	var allOldFiles []string
	var addedFiles []string
	var untrackedFiles []string
	var modifiedFiles []string
	// renamedFiles := make(map[string]string) // rename to -> rename from

	var renamedFilesWithSummaryGroup string
	groups := strings.Split(res, "\r\r\r\r\r\r")
	for _, group := range groups {
		idx := strings.Index(group, ":")
		if idx <= 0 {
			continue
		}
		groupName := group[:idx]
		groupContent := strings.TrimSpace(group[idx+1:])
		switch groupName {
		case "all-new-files":
			allNewFiles = splitLinesFilterEmpty(groupContent)
		case "all-old-files":
			allOldFiles = splitLinesFilterEmpty(groupContent)
		case "added-files":
			addedFiles = splitLinesFilterEmpty(groupContent)
		case "untracked-files":
			untrackedFiles = splitLinesFilterEmpty(groupContent)
		case "modified-files":
			modifiedFiles = splitLinesFilterEmpty(groupContent)
		case "renamed-files-with-summary":
			renamedFilesWithSummaryGroup = groupContent
			// case "renamed-files":
			// 	lines := splitLinesFilterEmpty(groupContent)
			// 	if len(lines)%2 != 0 {
			// 		err = fmt.Errorf("internal error, expect git return rename pairs, found:%d", len(lines))
			// 		return
			// 	}
			// 	renamedFiles = make(map[string]string, len(lines)/2)
			// 	for i := 0; i < len(lines); i += 2 {
			// 		from := strings.TrimPrefix(lines[i], "rename from ")
			// 		to := strings.TrimPrefix(lines[i+1], "rename to ")

			// 		renamedFiles[to] = from
			// 	}
		}
	}

	fileDetailsMap = make(map[string]*FileDetail, len(allNewFiles))
	for _, file := range allNewFiles {
		fileDetailsMap[file] = &FileDetail{}
	}
	for _, file := range allOldFiles {
		_, ok := fileDetailsMap[file]
		if !ok {
			// deleted files has two type of changes:
			// 1.deleted
			// 2.renamed to another file
			// an example:
			//   HEAD:   ab.txt
			// HEAD~1:   a.txt b.txt
			// results:  ab.txt: renamedFrom=a.txt,contentChanged=true
			//           a.txt:  deleted = true
			//           b.txt:  deleted = true
			fileDetailsMap[file] = &FileDetail{
				Deleted: true, // it may be renmaed also
			}
		}
	}
	// add untracked files
	for _, file := range untrackedFiles {
		fd := fileDetailsMap[file]
		if fd != nil {
			if fd.Deleted {
				// untracked file in allOld, mark as update
				fd.Deleted = false
				fd.ContentChanged = true
				continue
			}
			// untracked file in allNewFiles
			fd.IsNew = true
			continue
		}
		// neither in allNew nor allOld, mark as new
		fileDetailsMap[file] = &FileDetail{
			IsNew: true,
		}
	}

	for _, file := range addedFiles {
		fileDetailsMap[file].IsNew = true
	}
	// NOTE: this only contains un-renamed updates
	for _, file := range modifiedFiles {
		fileDetailsMap[file].ContentChanged = true
	}

	// check renames and updates
	// iterate for each 'rename ' at line begin
	// do not consider filename containing newline,"{" and "}", and space in the end or start
	parseRenames(renamedFilesWithSummaryGroup, func(newFile, oldFile, percent string) {
		d := fileDetailsMap[newFile]

		d.RenamedFrom = oldFile
		d.ContentChanged = d.ContentChanged || percent != "100%"
	})

	return
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
