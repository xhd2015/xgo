package main

import (
	"strings"
	"testing"
)

func TestHelpText(t *testing.T) {
	if helpText == "" {
		t.Error("helpText is empty")
	}
	if !strings.Contains(helpText, "Usage:") {
		t.Error("helpText should contain 'Usage:'")
	}
	if !strings.Contains(helpText, "list") {
		t.Error("helpText should mention 'list' command")
	}
	if !strings.Contains(helpText, "download") {
		t.Error("helpText should mention 'download' command")
	}
}

func TestRunHelp(t *testing.T) {
	err := run([]string{"--help"})
	if err != nil {
		t.Errorf("unexpected error for --help: %v", err)
	}

	err = run([]string{"-h"})
	if err != nil {
		t.Errorf("unexpected error for -h: %v", err)
	}
}

func TestRunDownloadHelp(t *testing.T) {
	err := run([]string{"download", "--help"})
	if err != nil {
		t.Errorf("unexpected error for 'download --help': %v", err)
	}

	err = run([]string{"download", "-h"})
	if err != nil {
		t.Errorf("unexpected error for 'download -h': %v", err)
	}
}

func TestRunNoArgs(t *testing.T) {
	err := run(nil)
	if err == nil {
		t.Fatal("expected error for nil args")
	}
	if !strings.Contains(err.Error(), "requires cmd") {
		t.Errorf("expected 'requires cmd' error, got: %v", err)
	}

	err = run([]string{})
	if err == nil {
		t.Fatal("expected error for empty args")
	}
	if !strings.Contains(err.Error(), "requires cmd") {
		t.Errorf("expected 'requires cmd' error, got: %v", err)
	}
}

func TestRunUnrecognizedCmd(t *testing.T) {
	err := run([]string{"unknown-cmd"})
	if err == nil {
		t.Fatal("expected error for unrecognized cmd")
	}
	if !strings.Contains(err.Error(), "unrecognized cmd") {
		t.Errorf("expected 'unrecognized cmd' error, got: %v", err)
	}
}

func TestParseDownloadArgsUnrecognizedFlag(t *testing.T) {
	_, _, err := parseDownloadArgs([]string{"--unknown"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(err.Error(), "unrecognized flag") {
		t.Errorf("expected 'unrecognized flag' error, got: %v", err)
	}
}
