package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCfg  *config
		wantErr  string
	}{
		{
			name: "help flag",
			args: []string{"--help"},
			wantCfg: &config{event: "push", showHelp: true},
		},
		{
			name:    "workflow file only",
			args:    []string{"my-workflow.yml"},
			wantCfg: &config{workflowFile: "my-workflow.yml", event: "push"},
		},
		{
			name:    "with event",
			args:    []string{"--event", "pull_request", "my-workflow.yml"},
			wantCfg: &config{workflowFile: "my-workflow.yml", event: "pull_request"},
		},
		{
			name:    "with job",
			args:    []string{"--job", "tests", ".github/workflows/go.yml"},
			wantCfg: &config{workflowFile: ".github/workflows/go.yml", event: "push", job: "tests"},
		},
		{
			name:    "dry run",
			args:    []string{"-n", "my-workflow.yml"},
			wantCfg: &config{workflowFile: "my-workflow.yml", event: "push", dryRun: true},
		},
		{
			name:    "dry run long flag",
			args:    []string{"--dry-run", "my-workflow.yml"},
			wantCfg: &config{workflowFile: "my-workflow.yml", event: "push", dryRun: true},
		},
		{
			name:    "list flag",
			args:    []string{"--list", "my-workflow.yml"},
			wantCfg: &config{workflowFile: "my-workflow.yml", event: "push", list: true},
		},
		{
			name:    "all flags combined",
			args:    []string{"--event", "pull_request", "--job", "build", "--dry-run", "ci.yml"},
			wantCfg: &config{workflowFile: "ci.yml", event: "pull_request", job: "build", dryRun: true},
		},
		{
			name:    "extra args via --",
			args:    []string{"ci.yml", "--", "-v", "--rm"},
			wantCfg: &config{workflowFile: "ci.yml", event: "push"},
		},
		{
			name:    "missing workflow file",
			args:    []string{},
			wantCfg: &config{event: "push"},
		},
		{
			name:    "missing event value",
			args:    []string{"--event"},
			wantErr: "--event requires a value",
		},
		{
			name:    "missing job value",
			args:    []string{"--job"},
			wantErr: "--job requires a value",
		},
		{
			name:    "unexpected positional arg",
			args:    []string{"a.yml", "b.yml"},
			wantErr: "unexpected positional argument: b.yml",
		},
		{
			name:    "unrecognized flag",
			args:    []string{"--unknown", "a.yml"},
			wantErr: "unrecognized flag: --unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := parseArgs(tt.args)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.workflowFile != tt.wantCfg.workflowFile {
				t.Errorf("workflowFile: got %q, want %q", cfg.workflowFile, tt.wantCfg.workflowFile)
			}
			if cfg.event != tt.wantCfg.event {
				t.Errorf("event: got %q, want %q", cfg.event, tt.wantCfg.event)
			}
			if cfg.job != tt.wantCfg.job {
				t.Errorf("job: got %q, want %q", cfg.job, tt.wantCfg.job)
			}
			if cfg.dryRun != tt.wantCfg.dryRun {
				t.Errorf("dryRun: got %v, want %v", cfg.dryRun, tt.wantCfg.dryRun)
			}
			if cfg.list != tt.wantCfg.list {
				t.Errorf("list: got %v, want %v", cfg.list, tt.wantCfg.list)
			}
			if cfg.showHelp != tt.wantCfg.showHelp {
				t.Errorf("showHelp: got %v, want %v", cfg.showHelp, tt.wantCfg.showHelp)
			}
		})
	}
}

func TestParseArgsExtraArgs(t *testing.T) {
	_, extra, err := parseArgs([]string{"ci.yml", "--", "-v", "--rm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(extra) != 2 || extra[0] != "-v" || extra[1] != "--rm" {
		t.Errorf("extra args: got %v, want [-v --rm]", extra)
	}
}

func TestEnsureActAlreadyInstalled(t *testing.T) {
	runner := &actRunner{
		goos:     "darwin",
		lookPath: func(s string) (string, error) { return "/usr/local/bin/act", nil },
		runCmd:   nil,
	}
	if err := runner.ensureAct(); err != nil {
		t.Fatalf("expected no error when act is found, got: %v", err)
	}
}

func TestEnsureActAutoInstallOnDarwin(t *testing.T) {
	var calledBrew bool
	runner := &actRunner{
		goos: "darwin",
		lookPath: func(s string) (string, error) {
			return "", fmt.Errorf("not found")
		},
		runCmd: func(name string, args ...string) error {
			if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "act" {
				calledBrew = true
			}
			return nil
		},
	}
	if err := runner.ensureAct(); err != nil {
		t.Fatalf("expected no error on darwin auto-install, got: %v", err)
	}
	if !calledBrew {
		t.Error("expected brew install act to be called")
	}
}

func TestEnsureActBrewInstallFails(t *testing.T) {
	runner := &actRunner{
		goos: "darwin",
		lookPath: func(s string) (string, error) {
			return "", fmt.Errorf("not found")
		},
		runCmd: func(name string, args ...string) error {
			return fmt.Errorf("brew failed")
		},
	}
	err := runner.ensureAct()
	if err == nil {
		t.Fatal("expected error when brew install fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to install act via brew") {
		t.Errorf("expected brew install error, got: %v", err)
	}
}

func TestEnsureActNotDarwinError(t *testing.T) {
	runner := &actRunner{
		goos: "linux",
		lookPath: func(s string) (string, error) {
			return "", fmt.Errorf("not found")
		},
		runCmd: nil,
	}
	err := runner.ensureAct()
	if err == nil {
		t.Fatal("expected error on non-darwin when act not found, got nil")
	}
	if !strings.Contains(err.Error(), "act is not installed") {
		t.Errorf("expected 'act is not installed' error, got: %v", err)
	}
}

func TestGoosFromXgoTestGoos(t *testing.T) {
	v := resolveGoosFromEnv("linux", "")
	if v != "linux" {
		t.Errorf("expected XGO_TEST_GOOS to take precedence, got %q", v)
	}
}

func TestGoosFromGoosEnv(t *testing.T) {
	v := resolveGoosFromEnv("", "windows")
	if v != "windows" {
		t.Errorf("expected GOOS env to be used, got %q", v)
	}
}

func resolveGoosFromEnv(xgoTestGoos, goosEnv string) string {
	if xgoTestGoos != "" {
		return xgoTestGoos
	}
	if goosEnv != "" {
		return goosEnv
	}
	return ""
}
