package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/model"
)

// git blame --porcelain desgined for machine consume
// --porcelain format:
//   header:        commit from-line to-line subsequent-ones
//   header-fields  prop value
//                  ...
//                  <TAB> content
//
// example:
//
// 2636b0147422024ae81b3c3809b3b90d34c35502 1 1 5
// author Some One
// author-mail <some@some.com>
// author-time 1662995182
// author-tz +0800
// committer Huadong Xiao
// committer-mail <some@some.com>
// committer-time 1662995688
// committer-tz +0800
// summary init
// boundary
// filename go.mod
// 	module some content
// 2636b0147422024ae81b3c3809b3b90d34c35502 2 2

// 2636b0147422024ae81b3c3809b3b90d34c35502 3 3
// 	go 1.14
// 2636b0147422024ae81b3c3809b3b90d34c35502 4 4

// 2636b0147422024ae81b3c3809b3b90d34c35502 5 5
// 	require (
// 1d679aaff5c670786d68a651882f1884e0501d68 6 6 1
//

type parseBlameState int

const (
	parseBlameState_Header parseBlameState = iota
	parseBlameState_Props  parseBlameState = iota
)

func Blame(dir string, ref string, file string) (lines []*model.BlameInfo, commits map[string]*model.BlameCommit, err error) {
	// HEAD makes no sense here
	if ref == "" {
		err = fmt.Errorf("requires ref")
		return
	}
	if file == "" {
		err = fmt.Errorf("requires file")
		return
	}
	res, err := cmd.Dir(dir).Output("git", "blame", "--porcelain", ref, "--", file)
	if err != nil {
		return
	}
	resLines := strings.Split(res, "\n")

	commits = make(map[string]*model.BlameCommit)
	state := parseBlameState_Header
	var lineNum int64
	var commitHash string
	var newCommit *model.BlameCommit
	for _, line := range resLines {
		if strings.HasPrefix(line, "\t") {
			if commitHash != "" && newCommit != nil {
				commits[commitHash] = newCommit
			}
			commitHash = ""
			newCommit = nil
			lineNum = 0
			state = parseBlameState_Header
			continue
		}
		switch state {
		case parseBlameState_Header:
			headSplits := strings.Split(line, " ")
			if len(headSplits) < 3 {
				err = fmt.Errorf("line %s: header fields less than 3", line)
				return
			}
			commitHash = headSplits[0]
			if commitHash == "" {
				err = fmt.Errorf("line %s: header empty commit hash", line)
				return
			}
			// [0]: commit hash
			// [1]: old line
			// [2]: new line (of current commit)
			// [3]: number of subsequent lines with same commit (optional)
			lineNum, err = strconv.ParseInt(headSplits[2], 10, 64)
			if err != nil {
				err = fmt.Errorf("line %s: header field line: %v", line, err)
				return
			}
			_, ok := commits[commitHash]
			if !ok {
				newCommit = &model.BlameCommit{
					Commit: &model.Commit{
						Hash: commitHash,
					},
				}
			}
			lines = append(lines, &model.BlameInfo{
				Line:       lineNum,
				CommitHash: commitHash,
			})
			state = parseBlameState_Props
		case parseBlameState_Props:
			if newCommit == nil {
				continue
			}
			propValue := strings.SplitN(line, " ", 2)
			if len(propValue) == 0 {
				continue
			}
			prop := propValue[0]
			var value string
			if len(propValue) > 1 {
				value = propValue[1]
			}
			switch prop {
			case "boundary":
				newCommit.Boundary = true
			case "summary":
				// TODO what if summary across multiple lines?
				newCommit.Msg = value
			case "author-mail":
				newCommit.AuthorEmail = value
			case "author":
				newCommit.AuthorName = value
			case "author-time":
				time, _ := strconv.ParseInt(value, 10, 64)
				newCommit.AuthorTimestamp = time
			}
		default:
			err = fmt.Errorf("unknown state: %v", state)
			return
		}
	}
	// end
	if commitHash != "" && newCommit != nil {
		commits[commitHash] = newCommit
	}
	return
}

// -t: show timestamp instead of datetime
// git blame -t --show-email output example:
// ^c9e7754 (<some@some.com> 1662996314 +0800 1) const rootElement = document.getElementById('root');
// ^c9e7754 (<some@some.com> 1662996314 +0800 2) const root = createRoot(rootElement)
// 5e0ea593 (<some@some.com> 1681291891 +0800 3) root.render(React.createElement(routes));

// ^c9e7754 means the initial commit(boundary)
func BlamePlain(dir string, ref string, file string) ([]*model.PlainBlameInfo, error) {
	// HEAD makes no sense here
	if ref == "" {
		return nil, fmt.Errorf("requires ref")
	}
	if file == "" {
		return nil, fmt.Errorf("requires file")
	}

	res, err := cmd.Dir(dir).Output("git", "blame", "-t", "--show-email", ref, "--", file)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(res, "\n")
	blameInfoList := make([]*model.PlainBlameInfo, 0, len(lines))
	for _, line := range lines {
		blameInfo, err := parsePlainBlame(line)
		if err != nil {
			return nil, fmt.Errorf("line %s: %v", line, err)
		}
		blameInfoList = append(blameInfoList, blameInfo)
	}
	return blameInfoList, nil
}

func parsePlainBlame(line string) (*model.PlainBlameInfo, error) {
	spaceIdx := strings.Index(line, " (<")
	if spaceIdx < 0 {
		return nil, fmt.Errorf("blame prop mark not found")
	}
	commit := line[:spaceIdx]
	var boundary bool
	if strings.HasPrefix(commit, "^") {
		boundary = true
		commit = commit[1:]
	}

	line = line[spaceIdx+3:]

	propEndIdx := strings.Index(line, ") ")
	if propEndIdx < 0 {
		return nil, fmt.Errorf("blame prop mark end not found")
	}
	props := line[:propEndIdx]

	emailEndIdx := strings.Index(props, "> ")
	if emailEndIdx < 0 {
		return nil, fmt.Errorf("blame email end mark not found")
	}

	email := props[:emailEndIdx]

	props = props[emailEndIdx+2:]
	tsEndIdx := strings.Index(props, " ")
	if tsEndIdx < 0 {
		return nil, fmt.Errorf("blame timestamp end mark not found")
	}

	timestampStr := props[:tsEndIdx]
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("blame parsing timestamp: %v", err)
	}
	props = props[tsEndIdx+1:]

	lineStartIdx := strings.LastIndex(props, " ")
	if lineStartIdx < 0 {
		return nil, fmt.Errorf("blame line start mark not found")
	}
	lineNum, err := strconv.ParseInt(props[lineStartIdx+1:], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("blame parsing line: %v", err)
	}

	return &model.PlainBlameInfo{
		Line:        lineNum,
		CommitHash:  commit,
		AuthorEmail: email,
		Boundary:    boundary,
		Timestamp:   timestamp,
	}, nil
}
