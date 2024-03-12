package main

import (
	"fmt"
	"strings"
)

type options struct {
	flagA      bool
	projectDir string
	output     string
	verbose    bool
	xgoSrc     string
	debug      string
	vscode     string
	withGoroot string
	dumpIR     string

	// TODO: make these options available only at develop
	syncXgoOnly  bool
	syncWithLink bool

	remainArgs []string
}

func parseOptions(args []string) (*options, error) {
	var flagA bool
	var verbose bool
	var projectDir string
	var output string
	var debug string
	var vscode string

	var xgoSrc string
	var syncXgoOnly bool
	var syncWithLink bool
	var withGoroot string
	var dumpIR string

	var remainArgs []string
	nArg := len(args)
	for i := 0; i < nArg; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			remainArgs = append(remainArgs, arg)
			continue
		}
		if arg == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if arg == "-a" {
			flagA = true
			continue
		}
		if arg == "-v" {
			verbose = true
			continue
		}
		if arg == "--sync-xgo-only" {
			syncXgoOnly = true
			continue
		}
		if arg == "--sync-with-link" {
			syncWithLink = true
			continue
		}
		ok, err := tryParseFlagValue("--project-dir", &projectDir, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		ok, err = tryParseFlagsValue([]string{"-o", "--output"}, &output, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		ok, err = tryParseFlagsValue([]string{"--debug"}, &debug, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		ok, err = tryParseFlagsValue([]string{"--vscode"}, &vscode, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		ok, err = tryParseFlagsValue([]string{"--xgo-src"}, &xgoSrc, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		ok, err = tryParseFlagsValue([]string{"--with-goroot"}, &withGoroot, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		ok, err = tryParseFlagsValue([]string{"--dump-ir"}, &dumpIR, &i, args)
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		return nil, fmt.Errorf("unrecognized flag:%s", arg)
	}

	return &options{
		flagA:        flagA,
		verbose:      verbose,
		projectDir:   projectDir,
		output:       output,
		xgoSrc:       xgoSrc,
		debug:        debug,
		vscode:       vscode,
		withGoroot:   withGoroot,
		dumpIR:       dumpIR,
		syncXgoOnly:  syncXgoOnly,
		syncWithLink: syncWithLink,
		remainArgs:   remainArgs,
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
