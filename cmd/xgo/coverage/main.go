package coverage

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/xgo/support/coverage"
)

func Main(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "requires cmd\n")
		os.Exit(1)
	}
	cmd := args[0]
	args = args[1:]
	if cmd == "help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return
	}
	if cmd != "merge" && cmd != "compact" && cmd != "serve" {
		fmt.Fprintf(os.Stderr, "unrecognized cmd: %s\n", cmd)
		return
	}
	if cmd == "serve" {
		err := handleServe(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
			return
		}
		return
	}

	var remainArgs []string
	var outFile string
	n := len(args)

	var flagHelp bool
	var excludePrefix []string
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if arg == "-o" {
			if i+1 >= n {
				fmt.Fprintf(os.Stderr, "%s requires file\n", arg)
				os.Exit(1)
			}
			outFile = args[i+1]
			i++
			continue
		}
		if arg == "--exclude-prefix" {
			if i+1 >= n {
				fmt.Fprintf(os.Stderr, "%s requires argument\n", arg)
				os.Exit(1)
			}
			if args[i+1] == "" {
				fmt.Fprintf(os.Stderr, "%s requires non empty argument\n", arg)
				os.Exit(1)
			}
			excludePrefix = append(excludePrefix, args[i+1])
			i++
			continue
		}
		if arg == "--help" || arg == "-h" {
			flagHelp = true
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			remainArgs = append(remainArgs, arg)
			continue
		}
		fmt.Fprintf(os.Stderr, "unrecognized flag: %s\n", arg)
		os.Exit(1)
	}
	if flagHelp {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return
	}

	switch cmd {
	case "merge":
		if len(remainArgs) == 0 {
			fmt.Fprintf(os.Stderr, "requires files\n")
			os.Exit(1)
		}
		err := mergeCover(remainArgs, outFile, excludePrefix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	case "compact":
		if len(remainArgs) == 0 {
			fmt.Fprintf(os.Stderr, "requires file\n")
			os.Exit(1)
		}
		if len(remainArgs) != 1 {
			fmt.Fprintf(os.Stderr, "compact requires exactly 1, found: %v\n", remainArgs)
			os.Exit(1)
		}
		// compact is a special case of merge
		err := mergeCover(remainArgs, outFile, excludePrefix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unrecognized cmd: %s\n", cmd)
		os.Exit(1)
	}
}

func mergeCover(files []string, outFile string, excludePrefix []string) error {
	// var mode string
	covs := make([][]*coverage.CovLine, 0, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		_, cov := coverage.Parse(string(content))
		covs = append(covs, cov)
		// if mode == "" {
		// 	mode = covMode
		// }
	}
	res := coverage.Merge(covs...)
	res = coverage.Filter(res, func(line *coverage.CovLine) bool {
		return !hasAnyPrefix(line.Prefix, excludePrefix)
	})

	// if mode == "" {
	// 	mode = "set"
	// }
	// always set mode to count
	mergedCov := coverage.Format("count", res)

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

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
