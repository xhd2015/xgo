package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/model"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-inspect/sh"
)

func GrepLines(dir string, ref string, options *model.GrepLineOptions) (lines map[string][]int, err error) {
	if ref == "" {
		err = fmt.Errorf("requires ref")
		return
	}
	if options == nil || len(options.Patterns) == 0 {
		err = fmt.Errorf("require patterns")
		return
	}
	// git grep -n  -E  -o -i -e interestRate -e pmt origin/master -- '*.go'
	// NOTE: exit 1 if no match
	// -z --null: output path names seperated with \0, we don't need it
	//            just assume file names are  file
	// --full-name: when running in subdir, print file absolutely
	// -o: only print match part
	// -e PATTERN: next param is pattern
	// -E: posix regex
	// -n: show line number

	// example output:
	//
	// 	origin/master:src/biz/batch_query_items_instalment_info_biz.go:618:InterestRate
	//  origin/master:src/biz/batch_query_items_instalment_info_biz.go:683:InterestRate
	//  origin/master:src/biz/batch_query_items_instalment_info_biz.go:683:InterestRate
	var args []string
	if options.IgnoreCase {
		args = append(args, "-i")
	}
	if options.Posix {
		args = append(args, "-E")
	}
	if options.WordMatch {
		args = append(args, "-w")
	}
	args = append(args, "-o") // only match
	args = append(args, "-n") // line number
	args = append(args, "--full-name")
	patternCnt := 0
	for _, p := range options.Patterns {
		if p != "" {
			patternCnt++
			args = append(args, "-e")
			args = append(args, p)
		}
	}
	if patternCnt == 0 {
		err = fmt.Errorf("require non-empty patterns")
		return
	}

	if ref != COMMIT_WORKING {
		args = append(args, ref)
	}

	args = append(args, "--")
	args = append(args, options.Files...)
	arg := sh.Quotes(args...)

	res, err := RunCommand(dir, func(commands []string) []string {
		return append(commands, []string{
			fmt.Sprintf("git grep %s || true", arg),
		}...)
	})
	if err != nil {
		return nil, err
	}

	prefix := ref + ":"
	outputLines := splitLinesFilterEmpty(res)

	lines = make(map[string][]int)
	for _, line := range outputLines {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		fileLine := line[len(prefix):]
		split := strings.SplitN(fileLine, ":", 3)
		if len(split) < 3 {
			continue
		}
		file := split[0]
		lineNum, _ := strconv.ParseInt(split[1], 10, 64)
		// ignore remaining content

		if file == "" || lineNum <= 0 {
			continue
		}
		lines[file] = append(lines[file], int(lineNum))
	}
	return
}
