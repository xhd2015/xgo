package exec_tool

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/flag"
)

type options struct {
	// by default all compiler's extra function
	// are disabled, unless explicitly called with enable
	enable     bool
	logCompile bool
	debug      string

	debugWithDlv bool

	testCompile bool // --test-compile

	remainArgs []string
}

func parseOptions(args []string, stopAfterFirstArg bool) (*options, error) {
	var enable bool
	var verbose bool
	var debug string

	var testCompile bool
	var debugWithDlv bool

	var remainArgs []string

	nArg := len(args)
	for i := 0; i < nArg; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			if stopAfterFirstArg {
				remainArgs = append(remainArgs, args[i+1:]...)
				break
			} else {
				remainArgs = append(remainArgs, arg)
				continue
			}
		}
		if arg == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if arg == "--enable" {
			enable = true
			continue
		}

		if arg == "--log-compile" {
			verbose = true
			continue
		}
		if arg == "--test-compile" {
			testCompile = true
			continue
		}
		if arg == "--debug-with-dlv" {
			debugWithDlv = true
			continue
		}

		ok, err := flag.TryParseFlagsValue([]string{"--debug"}, &debug, nil, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		return nil, fmt.Errorf("unrecognized flag:%s", arg)
	}

	return &options{
		enable:     enable,
		logCompile: verbose,
		debug:      debug,

		testCompile:  testCompile,
		debugWithDlv: debugWithDlv,

		remainArgs: remainArgs,
	}, nil
}
