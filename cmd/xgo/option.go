package main

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/flag"
)

type options struct {
	flagA      bool
	projectDir string
	output     string
	flagV      bool
	flagX      bool
	flagC      bool
	xgoSrc     string
	debug      string
	vscode     string
	withGoroot string
	dumpIR     string

	logCompile bool

	logDebug *string

	noBuildOutput   bool
	noInstrument    bool
	resetInstrument bool
	noSetup         bool

	// TODO: make these options available only at develop
	// deprecated
	syncXgoOnly   bool
	setupDev      bool
	buildCompiler bool
	syncWithLink  bool

	// recognize go flags as is
	// -gcflags
	gcflags string

	remainArgs []string
}

func parseOptions(args []string) (*options, error) {
	var flagA bool
	var flagV bool
	var flagX bool
	var flagC bool // -c: used by go test
	var projectDir string
	var output string
	var debug string
	var vscode string

	var noInstrument bool
	var resetInstrument bool
	var noSetup bool

	var xgoSrc string
	var syncXgoOnly bool
	var setupDev bool
	var buildCompiler bool

	var syncWithLink *bool
	var withGoroot string
	var dumpIR string

	var logCompile bool
	var logDebug *string

	var noBuildOutput bool

	var gcflags string

	var remainArgs []string
	nArg := len(args)

	type FlagValue struct {
		Flags []string
		Value *string
	}

	var flagValues []FlagValue = []FlagValue{
		{
			Flags: []string{"--project-dir"},
			Value: &projectDir,
		},
		{
			Flags: []string{"-o"},
			Value: &output,
		},
		{
			Flags: []string{"--debug"},
			Value: &debug,
		},
		{
			Flags: []string{"--vscode"},
			Value: &vscode,
		},
		{
			Flags: []string{"--xgo-src"},
			Value: &xgoSrc,
		},
		{
			Flags: []string{"--with-goroot"},
			Value: &withGoroot,
		},
		{
			Flags: []string{"--dump-ir"},
			Value: &dumpIR,
		},
		{
			Flags: []string{"-gcflags"},
			Value: &gcflags,
		},
	}
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
		if arg == "-" {
			return nil, fmt.Errorf("unrecognized flag:%s", arg)
		}
		if arg == "-a" {
			flagA = true
			continue
		}
		if arg == "-x" {
			flagX = true
			continue
		}
		if arg == "-c" {
			flagC = true
			continue
		}
		if arg == "-v" {
			flagV = true
			continue
		}
		if arg == "--log-compile" {
			logCompile = true
			continue
		}
		if arg == "--sync-xgo-only" {
			syncXgoOnly = true
			continue
		}
		if arg == "--setup-dev" {
			setupDev = true
			continue
		}
		if arg == "--build-compiler" {
			buildCompiler = true
			continue
		}
		if arg == "--sync-with-link" {
			v := true
			syncWithLink = &v
			continue
		}
		if arg == "--no-build-output" {
			noBuildOutput = true
			continue
		}
		if arg == "--no-instrument" {
			noInstrument = true
			continue
		}
		if arg == "--reset-instrument" {
			resetInstrument = true
			continue
		}
		if arg == "--no-setup" {
			noSetup = true
			continue
		}
		if arg == "--log-debug" {
			var logDebugStr string
			idx := strings.Index(arg, "=")
			if idx >= 0 {
				logDebugStr = arg[idx+1:]
			}
			logDebug = &logDebugStr
			continue
		}
		var found bool
		for _, flagVal := range flagValues {
			ok, err := flag.TryParseFlagsValue(flagVal.Flags, flagVal.Value, &i, args)
			if err != nil {
				return nil, err
			}
			if ok {
				found = true
				break
			}
		}
		if found {
			continue
		}

		// check if single dash flags, this is usually go flags, such as -ldflags...
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			eqIdx := strings.Index(arg, "=")
			if eqIdx >= 0 {
				// things like -count=0
				remainArgs = append(remainArgs, arg)
				continue
			}
			if i+1 < nArg && !strings.HasPrefix(args[i+1], "-") {
				// -count 0
				// check if next arg starts with "-" (i.e. an option)
				remainArgs = append(remainArgs, arg, args[i+1])
				i++
				continue
			}
			// -count
			remainArgs = append(remainArgs, arg)
			continue
		}

		return nil, fmt.Errorf("unrecognized flag:%s", arg)
	}

	return &options{
		flagA:      flagA,
		flagV:      flagV,
		flagX:      flagX,
		flagC:      flagC,
		projectDir: projectDir,
		output:     output,
		xgoSrc:     xgoSrc,
		debug:      debug,
		vscode:     vscode,
		withGoroot: withGoroot,
		dumpIR:     dumpIR,

		logCompile: logCompile,
		logDebug:   logDebug,

		noBuildOutput:   noBuildOutput,
		noInstrument:    noInstrument,
		resetInstrument: resetInstrument,
		noSetup:         noSetup,

		syncXgoOnly:   syncXgoOnly,
		setupDev:      setupDev,
		buildCompiler: buildCompiler,
		// default true
		syncWithLink: syncWithLink == nil || *syncWithLink,

		gcflags: gcflags,

		remainArgs: remainArgs,
	}, nil
}
