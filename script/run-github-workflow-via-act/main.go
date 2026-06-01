package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

const help = `Usage: run-github-workflow-via-act [options] <workflow.yml>

Run a GitHub Actions workflow locally via act.

Options:
  --event <event>  GitHub event type (push, pull_request). Default: push
  --job <job>      Run a specific job
  -n, --dry-run    Dry run (passes -n to act)
  --list           List jobs in the workflow (passes -l to act)
  --help           Show this help
  --               Pass remaining args directly to act
`

func main() {
	args := os.Args[1:]
	cfg, extraArgs, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if cfg.showHelp {
		fmt.Print(help)
		os.Exit(0)
	}
	if cfg.workflowFile == "" {
		fmt.Fprintf(os.Stderr, "missing required argument: workflow file\n\n")
		fmt.Fprint(os.Stderr, help)
		os.Exit(1)
	}

	runner := &actRunner{
		goos:     goos,
		lookPath: exec.LookPath,
		runCmd: func(name string, args ...string) error {
			return cmd.Run(name, args...)
		},
	}

	if err := runner.ensureAct(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	actArgs := []string{"-W", cfg.workflowFile}
	if cfg.list {
		actArgs = append(actArgs, "-l")
	} else {
		actArgs = append(actArgs, "--event", cfg.event)
	}
	if cfg.job != "" {
		actArgs = append(actArgs, "--job", cfg.job)
	}
	if cfg.dryRun {
		actArgs = append(actArgs, "-n")
	}
	actArgs = append(actArgs, extraArgs...)

	if err := cmd.Run("act", actArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "act failed: %v\n", err)
		os.Exit(1)
	}
}

func resolveGoos() string {
	if s := os.Getenv("XGO_TEST_GOOS"); s != "" {
		return s
	}
	if s := os.Getenv("GOOS"); s != "" {
		return s
	}
	return runtime.GOOS
}

var goos = resolveGoos()

type config struct {
	workflowFile string
	event        string
	job          string
	dryRun       bool
	list         bool
	showHelp     bool
}

func parseArgs(args []string) (*config, []string, error) {
	cfg := &config{
		event: "push",
	}
	var extraArgs []string
	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			extraArgs = append(extraArgs, args[i+1:]...)
			break
		}
		switch arg {
		case "--help":
			cfg.showHelp = true
			return cfg, nil, nil
		case "--event":
			i++
			if i >= n {
				return nil, nil, fmt.Errorf("--event requires a value")
			}
			cfg.event = args[i]
		case "--job":
			i++
			if i >= n {
				return nil, nil, fmt.Errorf("--job requires a value")
			}
			cfg.job = args[i]
		case "-n", "--dry-run":
			cfg.dryRun = true
		case "--list":
			cfg.list = true
		default:
			if !strings.HasPrefix(arg, "-") {
				if cfg.workflowFile != "" {
					return nil, nil, fmt.Errorf("unexpected positional argument: %s", arg)
				}
				cfg.workflowFile = arg
			} else {
				return nil, nil, fmt.Errorf("unrecognized flag: %s", arg)
			}
		}
	}
	return cfg, extraArgs, nil
}

type actRunner struct {
	goos     string
	lookPath func(string) (string, error)
	runCmd   func(string, ...string) error
}

func (r *actRunner) ensureAct() error {
	_, err := r.lookPath("act")
	if err == nil {
		return nil
	}
	if r.goos == "darwin" {
		fmt.Fprintf(os.Stderr, "act not found, installing via brew...\n")
		if err := r.runCmd("brew", "install", "act"); err != nil {
			return fmt.Errorf("failed to install act via brew: %w", err)
		}
		return nil
	}
	return fmt.Errorf("act is not installed; please install it first (https://github.com/nektos/act)")
}
