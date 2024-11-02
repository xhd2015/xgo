package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-inspect/sh"
	"github.com/xhd2015/xgo/support/cmd"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/model"
)

// git pretty format doc link: https://git-scm.com/docs/pretty-formats
// %H: commit hash
// %an: author name
// %ae: author email
// %at: author timestamp (unix)
// %ct: commit timestamp (unix)
// %s: subject (commit msg)
// %(describe): tag

// const commitFormat = `$'H=%H\r\rae=%ae\r\rat=%at\r\rct=%ct\r\rtag=%(describe)\r\rs=%s\r\r\r\r\r'`
const commitSplit = "\r\r\r\r\r\n"

func makeCommitFormat(needTag bool) string {
	tagClause := ""
	if needTag {
		tagClause = `tag=%(describe)\r\r`
	}
	return `$'H=%H\r\ran=%an\r\rae=%ae\r\rat=%at\r\rct=%ct\r\r` + tagClause + `s=%s\r\r\r\r\r'`
}

// git log --format='%H %s' A..B
//
//	H MSG
func ListCommits(dir string, beginRef string, ref string) ([]*model.Commit, error) {
	if ref == "" {
		return nil, fmt.Errorf("requires revision")
	}

	args := []string{"log", "--format=" + makeCommitFormat(true)}
	if beginRef != "" {
		// rev range
		args = append(args, "^"+beginRef, ref)
	} else {
		args = append(args, "ref")
	}
	res, err := cmd.Dir(dir).Output("git", args...)
	if err != nil {
		return nil, err
	}

	return parseCommits(res), nil
}

func splitBatch(refs []string, batch int) [][]string {
	if batch <= 0 {
		panic(fmt.Errorf("requires batch to be positive"))
	}
	n := len(refs)
	m := n / batch
	if n%batch != 0 {
		m++
	}
	batches := make([][]string, m)
	for i := 0; i < m; i++ {
		start := i * batch
		end := start + batch
		if end > n {
			end = n
		}
		batches[i] = refs[i*batch : end]
	}
	return batches
}

type GetCommitsOptions struct {
	Optional bool
}

// GetCommits get commits info by refs, the result is a mapping by ref name
func GetCommits(dir string, refs []string, opts ...GetCommitsOptions) (map[string]*model.Commit, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	var optional bool
	for _, opt := range opts {
		if opt.Optional {
			optional = true
		}
	}

	mappingByRef := make(map[string]*model.Commit, len(refs))
	var verifiedRefs []string
	if !optional {
		verifiedRefs = refs
	} else {
		verifiedRefs = make([]string, 0, len(refs))
		for _, ref := range refs {
			_, revErr := revParseVerified(dir, ref)
			if revErr != nil {
				if revErr == ErrNotExists {
					mappingByRef[ref] = &model.Commit{
						NotFound: true,
					}
					continue
				}
				return nil, revErr
			}
			verifiedRefs = append(verifiedRefs, ref)
		}
	}
	batches := splitBatch(verifiedRefs, 100)
	for _, batchRefs := range batches {
		cmds := make([]string, 0, len(batchRefs))
		for _, ref := range batchRefs {
			cmds = append(cmds, fmt.Sprintf("git log -1 --format=%s %s", makeCommitFormat(true), sh.Quote(ref)))
		}
		res, err := RunCommands(dir, cmds...)
		var convertBatchErr func(err error) error
		if !optional {
			convertBatchErr = func(err error) error {
				for _, ref := range batchRefs {
					_, revErr := revParseVerified(dir, ref)
					if revErr == ErrNotExists {
						return fmt.Errorf("%s does not exist or has been deleted", trimRef(ref))
					}
				}
				return err
			}
		} else {
			convertBatchErr = func(err error) error {
				return err
			}
		}

		if err != nil {
			return nil, convertBatchErr(err)
		}
		commits := parseCommits(res)
		if len(commits) != len(batchRefs) {
			return nil, convertBatchErr(fmt.Errorf("some ref is missing while getting commits"))
		}
		for i, commit := range commits {
			mappingByRef[batchRefs[i]] = commit
		}
	}
	return mappingByRef, nil
}

func GetCommit(dir string, ref string) (*model.Commit, error) {
	if ref == "" {
		return nil, fmt.Errorf("requires ref")
	}
	res, err := RunCommand(dir, func(commands []string) []string {
		return append(commands, fmt.Sprintf("git log -1 --format=%s %s", makeCommitFormat(true), sh.Quote(ref)))
	})
	if err != nil {
		return nil, convertRefError(dir, ref, err)
	}
	commits := parseCommits(res)
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commit")
	}
	return commits[0], nil
}

func ResolveDiffCommit(dir string, head string, base string) (exists bool, merged bool, headCommit string, baseCommit string, err error) {
	exists, merged, headCommit, baseCommit, _, err = doListRelativeToBase(dir, head, base, true)
	return
}

func ListCommitRelativeToBase(dir string, head string, base string) (exists bool, merged bool, commits []*model.Commit, err error) {
	exists, merged, _, _, commits, err = doListRelativeToBase(dir, head, base, false)
	return
}

// if either branch does not exists, exists is false.
// if merged, return an empty list
// if not merged, return a commit list from 'merge-base of head and base'(exclusive) to head
func doListRelativeToBase(dir string, head string, base string, forDiff bool) (exists bool, merged bool, headCommit string, baseCommit string, commits []*model.Commit, err error) {
	if head == "" {
		err = fmt.Errorf("requires head")
		return
	}
	if base == "" {
		err = fmt.Errorf("requires base")
		return
	}
	forDiffEnv := "false"
	if forDiff {
		forDiffEnv = "true"
	}
	useV2 := true
	// note, headCommit will always be resolve of head
	res, err := RunCommand(dir, func(commands []string) []string {
		commands = append(commands,
			fmt.Sprintf("forDiff=%s", forDiffEnv),
			fmt.Sprintf("head=$(git rev-parse --verify --quiet %s^{commit} || true)", sh.Quote(head)),
			fmt.Sprintf("base=$(git rev-parse --verify --quiet %s^{commit} || true)", sh.Quote(base)),
			`if [[ -z $head || -z $base ]];then echo "not_exists";exit;fi`,
			`mergeBase=$(git merge-base "$head" "$base")`,
			`if [[ $mergeBase = $head ]]; then `,
			`   echo "merged"`,
		)

		if useV2 {
			commands = append(commands,
				// assume the base always require --no-ff, all commits are merge points
				// the `--merges --reverse` find all merge points that belong to base unqiuely,
				// from earlies to latest.
				// then use `head -n1` to take the earlies one.
				`  if [[ $head != $base ]];then`,
				`    mergePoint=$(git rev-list --merges  --first-parent --reverse "${head}..${base}"|head -n1 || true)`,
				`  else`,
				// if head==base,that means they are just the same
				// just diff with previous commit
				`    mergePoint=$head`,
				`  fi`,
			)
		} else {
			// old implementation
			commands = append(commands,
				// find the merge point

				// when current head is merged into base, it has 2 possible cases:
				// - head is a merge point: we should return (head^2, head^1) as the compare pair, so format should be merge-base of head^2 and head^1 to head^2, including head.
				// - head is not a merge point: that means head's last commit is merged into base, but head itself does not move.
				`  head2=$(git rev-parse --verify --quiet "${head}^2" || true)`, // has ^2, so is a
				`  if [[ -n $head2 ]];then `,                                    // a merge point
				`       mergePoint=$head`,
				`  else`,
				// `       # echo "no bad";exit 1`,
				`       mergePoints=$(git rev-list --merges  --first-parent --reverse "${head}...${base}")`, // list merge commits reachable from only base, not head, this effectively finds all merges that lead head,
				// and we only need to find the first one that is child of head
				`       for commit in $mergePoints;do`,
				`          mbase=$(git merge-base $commit $head)`,
				`          if [[ $mbase = $head ]];then `,
				//             mbase is a merge point from head and base, so it
				//             is the same with its parent at ^2, which is from head,
				//             so we need to return mbase^1
				`              mergePoint=$commit`,
				`              break`,
				`          fi`,
				`       done`,
				`  fi`,
			)
		}
		commands = append(commands,
			// now that we have found the first mergePoint, we know the diff commit should be mergePoint^1
			// and branch's commits should be from branchStart to branch, with --first-parent
			// note for merge points: merge point is the same with its parent at ^2, but different with that at ^1
			`  if [[ -n ${mergePoint} ]];then`,
			`    preMergePoint=${mergePoint}^1`,
			`  fi`,
			`  baseCommit=$(git rev-parse --verify --quiet "${preMergePoint}" || true)`,
			`  branchStart=$(git merge-base "${baseCommit}" "${head}")`,
			`  format=${branchStart}...${head}`, // note: branch start is actually the parent of head, so ... will exclude those reachable from branchStart, leaving only those reacable from head, with --first-parent we can reserve only commit from head.
			``,
			`else `, // not merged
			`   echo "ok"`,
			`   baseCommit=$base`,
			`   format=${mergeBase}..${head}`,
			`fi`,
		)
		if forDiff {
			commands = append(commands, `echo "${head} ${baseCommit}"`)
		} else {
			commands = append(commands, fmt.Sprintf(`git log --first-parent --format=%s $format`, makeCommitFormat(true)))
		}
		return commands
	})
	if err != nil {
		return
	}
	firstLine := res
	idx := strings.Index(res, "\n")
	if idx >= 0 {
		firstLine = strings.TrimSpace(res[:idx])
		if idx+1 < len(res) {
			res = res[idx+1:]
		} else {
			res = ""
		}

	}
	if firstLine == "not_exists" {
		return
	}
	exists = true
	if firstLine == "merged" {
		merged = true
	}
	if forDiff {
		splits := strings.Split(strings.TrimSpace(res), " ")
		if len(splits) >= 2 {
			headCommit = splits[0]
			baseCommit = splits[1]
		}
	} else {
		commits = parseCommits(res)
	}
	return
}

func parseCommits(res string) []*model.Commit {
	lines := strings.Split(res, commitSplit)
	commits := make([]*model.Commit, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		splits := strings.Split(line, "\r\r")
		vars := make(map[string]string, 7)
		for _, kv := range splits {
			kvs := strings.SplitN(kv, "=", 2)
			if len(kvs) == 0 {
				continue
			}
			var v string
			if len(kvs) >= 2 {
				v = kvs[1]
			}
			vars[kvs[0]] = v
		}
		hash := vars["H"]
		authorName := vars["an"]
		authorEmail := vars["ae"]
		authorTimestamp, _ := strconv.ParseInt(vars["at"], 10, 64)
		commitTimestamp, _ := strconv.ParseInt(vars["ct"], 10, 64)
		msg := vars["s"]
		tag := vars["tag"]

		if hash == "" {
			continue
		}
		commits = append(commits, &model.Commit{
			Hash:            hash,
			Msg:             msg,
			AuthorName:      authorName,
			AuthorEmail:     authorEmail,
			AuthorTimestamp: authorTimestamp,
			CommitTimestamp: commitTimestamp,
			Tag:             tag,
		})
	}
	return commits
}
