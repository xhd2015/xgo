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
	dumpAST    string

	logCompile bool

	logDebug *string

	noBuildOutput   bool
	noInstrument    bool
	resetInstrument bool
	noSetup         bool

	// dev only
	debugWithDlv bool
	xgoHome      string

	// TODO: make these options available only at develop
	// deprecated
	syncXgoOnly   bool
	setupDev      bool
	buildCompiler bool
	syncWithLink  bool

	mod string
	// recognize go flags as is
	// -gcflags
	// can repeat
	gcflags []string

	overlay string
	modfile string

	// --trap-stdlib
	trapStdlib bool

	// xgo test --trace

	// --strace, --strace=on, --strace=off
	// --stack-stackTrace, --stack-stackTrace=off, --stack-stackTrace=on
	// to be used in test mode
	stackTrace string

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

	var debugWithDlv bool
	var xgoHome string

	var xgoSrc string
	var syncXgoOnly bool
	var setupDev bool
	var buildCompiler bool

	var syncWithLink *bool
	var withGoroot string
	var dumpIR string
	var dumpAST string

	var logCompile bool
	var logDebug *string

	var noBuildOutput bool

	var mod string
	var gcflags []string
	var overlay string
	var modfile string
	var stackTrace string
	var trapStdlib bool

	var remainArgs []string
	nArg := len(args)

	type FlagValue struct {
		Flags  []string
		Value  *string
		Single bool
		Set    func(v string)
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
			Flags: []string{"--dump-ast"},
			Value: &dumpAST,
		},
		{
			Flags: []string{"-mod"},
			Set: func(v string) {
				mod = v
			},
		},
		{
			Flags: []string{"-gcflags"},
			Set: func(v string) {
				gcflags = append(gcflags, v)
			},
		},
		{
			Flags: []string{"-overlay"},
			Set: func(v string) {
				overlay = v
			},
		},
		{
			Flags: []string{"-modfile"},
			Set: func(v string) {
				modfile = v
			},
		},
		{
			Flags:  []string{"--log-debug"},
			Single: true,
			Set: func(v string) {
				logDebug = &v
			},
		},
	}

	if isDevelopment {
		flagValues = append(flagValues, FlagValue{
			Flags: []string{"--xgo-home"},
			Value: &xgoHome,
		})
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

		argVal, ok := parseStackTraceFlag(arg)
		if ok {
			stackTrace = argVal
			continue
		}

		if arg == "--trap-stdlib" {
			trapStdlib = true
			continue
		}

		if isDevelopment && arg == "--debug-with-dlv" {
			debugWithDlv = true
			continue
		}
		var found bool
		for _, flagVal := range flagValues {
			if flagVal.Single {
				flag, val := flag.TrySingleFlag(flagVal.Flags, arg)
				if flag != "" {
					flagVal.Set(val)
					found = true
					break
				}
				continue
			}
			ok, err := flag.TryParseFlagsValue(flagVal.Flags, flagVal.Value, flagVal.Set, &i, args)
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
		dumpAST:    dumpAST,

		logCompile: logCompile,
		logDebug:   logDebug,

		noBuildOutput:   noBuildOutput,
		noInstrument:    noInstrument,
		resetInstrument: resetInstrument,
		noSetup:         noSetup,
		debugWithDlv:    debugWithDlv,
		xgoHome:         xgoHome,

		syncXgoOnly:   syncXgoOnly,
		setupDev:      setupDev,
		buildCompiler: buildCompiler,
		// default true
		syncWithLink: syncWithLink == nil || *syncWithLink,

		mod:        mod,
		gcflags:    gcflags,
		overlay:    overlay,
		modfile:    modfile,
		stackTrace: stackTrace,
		trapStdlib: trapStdlib,

		remainArgs: remainArgs,
	}, nil
}

func parseStackTraceFlag(arg string) (string, bool) {
	var stackTracePrefix string
	if strings.HasPrefix(arg, "--strace") {
		stackTracePrefix = "--strace"
	} else if strings.HasPrefix(arg, "--stack-trace") {
		stackTracePrefix = "--stack-trace"
	}
	if stackTracePrefix == "" {
		return "", false
	}
	if len(arg) == len(stackTracePrefix) {
		return "on", true
	}
	if arg[len(stackTracePrefix)] != '=' {
		return "", false
	}
	val := arg[len(stackTracePrefix)+1:]
	if val == "" || val == "on" {
		return "on", true
	}
	if val == "off" {
		return "off", true
	}
	panic(fmt.Errorf("unrecognized value %s: %s, expects on|off", arg, val))
}

func getPkgArgs(args []string) []string {
	n := len(args)
	newArgs := make([]string, 0, n)
	for i := 0; i < n; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			// stop at first non-arg
			newArgs = append(newArgs, args[i:]...)
			break
		}
		if arg == "-args" {
			// go test -args: pass everything after to underlying program
			break
		}
		eqIdx := strings.Index(arg, "=")
		if eqIdx >= 0 {
			// self hosted arg
			continue
		}
		// make --opt equivalent with -opt
		if strings.HasPrefix(arg, "--") {
			arg = arg[1:]
		}
		switch arg {
		case "-a", "-n", "-race", "-masan", "-asan", "-cover", "-v", "-work", "-x", "-linkshared", "-buildvcs", // shared among build,test,run
			"-args", "-c", "-json": // -json for test
			// zero arg
		default:
			// 1 arg
			i++
		}
	}
	return newArgs
}
