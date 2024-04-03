package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "requires cmd\n")
		os.Exit(1)
	}
	cmd := args[0]
	args = args[1:]

	var remainArgs []string
	var outFile string
	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "-o" {
			outFile = args[i+1]
			i++
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			remainArgs = append(remainArgs, arg)
			continue
		}
		fmt.Fprintf(os.Stderr, "unrecognized flag: %s\n", arg)
		os.Exit(1)
	}

	switch cmd {
	case "merge":
		if len(remainArgs) == 0 {
			fmt.Fprintf(os.Stderr, "requires files\n")
			os.Exit(1)
		}
		err := mergeCover(remainArgs, outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unrecognized cmd: %s\n", cmd)
		os.Exit(1)
	}
}

func mergeCover(files []string, outFile string) error {
	covs := make([][]*covLine, 0, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		covs = append(covs, parseCover(string(content)))
	}
	res := merge(covs)
	res = filter(res, func(line *covLine) bool {
		if strings.HasPrefix(line.prefix, "github.com/xhd2015/xgo/runtime/test") {
			return false
		}
		return true
	})

	mergedCov := formatCoverage("set", res)

	var out io.Writer = os.Stdout
	if outFile != "" {
		file, err := os.Create(outFile)
		if err != nil {
			return err
		}
		defer file.Close()
		out = file
	}
	_, err := io.WriteString(out, mergedCov)
	return err
}

func filter(covs []*covLine, check func(line *covLine) bool) []*covLine {
	n := len(covs)
	j := 0
	for i := 0; i < n; i++ {
		if check(covs[i]) {
			covs[j] = covs[i]
			j++
		}
	}
	return covs[:j]
}

func formatCoverage(mode string, lines []*covLine) string {
	strs := make([]string, 0, len(lines)+1)
	strs = append(strs, "mode: "+mode)
	for _, line := range lines {
		strs = append(strs, line.prefix+" "+strconv.FormatInt(line.count, 10))
	}
	return strings.Join(strs, "\n")
}

func merge(covs [][]*covLine) []*covLine {
	if len(covs) == 0 {
		return nil
	}
	if len(covs) == 1 {
		return covs[0]
	}
	result := covs[0]
	for i := 1; i < len(covs); i++ {
		result = mergeCov(result, covs[i])
	}
	return result
}

func mergeCov(a []*covLine, b []*covLine) []*covLine {
	for _, line := range b {
		idx := -1
		for i := 0; i < len(a); i++ {
			if a[i].prefix == line.prefix {
				idx = i
				break
			}
		}
		if idx < 0 {
			a = append(a, line)
		} else {
			// fmt.Printf("add %s %d %d\n", a[idx].prefix, a[idx].count, line.count)
			a[idx].count += line.count
		}
	}
	return a
}

type covLine struct {
	prefix string
	count  int64
}

func parseCover(content string) []*covLine {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "mode:") {
		lines = lines[1:]
	}
	covLines := make([]*covLine, 0, len(lines))
	for _, line := range lines {
		covLine := parseCovLine(line)
		if covLine == nil {
			continue
		}
		covLines = append(covLines, covLine)
	}
	return covLines
}

func parseCovLine(line string) *covLine {
	idx := strings.LastIndex(line, " ")
	if idx < 0 {
		return nil
	}
	cnt, err := strconv.ParseInt(line[idx+1:], 10, 64)
	if err != nil {
		cnt = 0
	}
	return &covLine{
		prefix: line[:idx],
		count:  cnt,
	}
}
