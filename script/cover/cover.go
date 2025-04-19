package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/xgo/support/coverage"
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
	covs := make([][]*coverage.CovLine, 0, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		_, lines := coverage.Parse(string(content))
		covs = append(covs, lines)
	}
	res := coverage.Merge(covs...)
	res = coverage.Filter(res, func(line *coverage.CovLine) bool {
		if strings.HasPrefix(line.Prefix, "github.com/xhd2015/xgo/runtime/test") {
			return false
		}
		return true
	})

	mergedCov := coverage.Format("set", res)

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
