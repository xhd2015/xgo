package main

import (
	"fmt"
	"strings"
)

type options struct {
	verbose    bool
	debug      string
	remainArgs []string
}

func parseOptions(args []string, stopAfterFirstArg bool) (*options, error) {
	var remainArgs []string
	var verbose bool
	var debug string

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

		if arg == "-v" {
			verbose = true
			continue
		}

		ok, err := tryParseFlagsValue([]string{"--debug"}, &debug, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		return nil, fmt.Errorf("unrecognized flag:%s", arg)
	}

	return &options{
		verbose:    verbose,
		debug:      debug,
		remainArgs: remainArgs,
	}, nil
}

func tryParseFlagValue(flag string, pval *string, pi *int, args []string) (ok bool, err error) {
	i := *pi
	val, next, ok := tryParseArg(flag, args[i])
	if !ok {
		return false, nil
	}
	if next {
		if i+1 >= len(args) {
			return false, fmt.Errorf("flag %s requires value", args[i])
		}
		val = args[i+1]
		*pi++
	}
	*pval = val
	return true, nil
}
func tryParseFlagsValue(flags []string, pval *string, pi *int, args []string) (ok bool, err error) {
	for _, flag := range flags {
		ok, err := tryParseFlagValue(flag, pval, pi, args)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func tryParseArg(flag string, arg string) (value string, next bool, ok bool) {
	if !strings.HasPrefix(arg, flag) {
		return "", false, false
	}
	if len(arg) == len(flag) {
		return "", true, true
	}
	if arg[len(flag)] == '=' {
		return arg[len(flag)+1:], false, true
	}
	return "", false, false
}
