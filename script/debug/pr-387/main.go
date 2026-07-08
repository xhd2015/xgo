package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const prNumber = "389"

var (
	ansiRe      = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	ciSuccessRe = regexp.MustCompile(`\s+success\s*https?://`)
	ciSkippedRe = regexp.MustCompile(`\s+skipped\s*https?://`)
	ciFailedRe  = regexp.MustCompile(`\s+(failure|cancelled|timed_out)\s*https?://`)
	ciPendingRe = regexp.MustCompile(`\s+(in_progress|queued|pending)\s*https?://`)
)

func outDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "cannot resolve script directory")
		os.Exit(2)
	}
	return filepath.Join(filepath.Dir(file), "out")
}

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runWithRetry(name string, retries int, delay time.Duration, args ...string) (string, error) {
	var out string
	var err error
	for attempt := 0; attempt <= retries; attempt++ {
		out, err = run(name, args...)
		if err == nil || !strings.Contains(out, "rate limit") {
			return out, err
		}
		if attempt < retries {
			time.Sleep(delay)
		}
	}
	return out, err
}

func inspect(check string, pass bool, evidence, reason string) {
	result := "FAIL"
	if pass {
		result = "PASS"
	}
	fmt.Printf("CHECK: %s\n", check)
	if evidence != "" {
		fmt.Printf("EVIDENCE: %s\n", evidence)
	}
	fmt.Printf("RESULT: %s\n", result)
	if !pass && reason != "" {
		fmt.Printf("REASON: %s\n", reason)
	}
	if !pass {
		os.Exit(1)
	}
}

type prView struct {
	State  string `json:"state"`
	Title  string `json:"title"`
	Status []struct {
		Name       string `json:"name"`
		Conclusion string `json:"conclusion"`
		Status     string `json:"status"`
	} `json:"statusCheckRollup"`
}

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func parseCIRuns(out string) (failed, pending, succeeded, skipped []string) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(stripANSI(line))
		if line == "" || strings.HasPrefix(line, "Workflow Runs") {
			continue
		}
		switch {
		case ciFailedRe.MatchString(line):
			failed = append(failed, line)
		case ciPendingRe.MatchString(line):
			pending = append(pending, line)
		case ciSuccessRe.MatchString(line):
			succeeded = append(succeeded, line)
		case ciSkippedRe.MatchString(line):
			skipped = append(skipped, line)
		}
	}
	return failed, pending, succeeded, skipped
}

func main() {
	_ = os.MkdirAll(outDir(), 0o755)

	// CHECK 1: github-fetch can read PR content
	fetchOut, fetchErr := run("github-fetch", "pr", prNumber)
	fetchPath := filepath.Join(outDir(), "github-fetch-pr.txt")
	_ = os.WriteFile(fetchPath, []byte(fetchOut), 0o644)
	inspect("github-fetch pr "+prNumber+" succeeds",
		fetchErr == nil && strings.Contains(fetchOut, "PR #"+prNumber),
		fetchPath,
		strings.TrimSpace(fetchOut))

	// CHECK 2: PR is open and about go1.26
	inspect("PR title mentions go1.26",
		strings.Contains(strings.ToLower(fetchOut), "go1.26"),
		strings.TrimSpace(firstLine(fetchOut)),
		"expected go1.26 in PR summary")

	// CHECK 2b: PR description credits omniaura contribution
	inspect("PR mentions omniaura/peyton attribution",
		strings.Contains(strings.ToLower(fetchOut), "omniaura") ||
			strings.Contains(strings.ToLower(fetchOut), "peyton"),
		fetchPath,
		"expected omniaura attribution in PR")

	// CHECK 3: github-fetch ci reports workflow status (primary; no gh auth required)
	ciOut, ciErr := runWithRetry("github-fetch", 3, 30*time.Second, "ci", prNumber)
	ciPath := filepath.Join(outDir(), "github-fetch-ci.txt")
	_ = os.WriteFile(ciPath, []byte(ciOut), 0o644)
	inspect("github-fetch ci returns workflow runs",
		ciErr == nil && strings.Contains(ciOut, "Workflow Runs"),
		ciPath,
		strings.TrimSpace(ciOut))

	failed, pending, succeeded, skipped := parseCIRuns(ciOut)
	summaryPath := filepath.Join(outDir(), "check-summary.txt")
	summary := fmt.Sprintf("succeeded: %d\nskipped: %d\nfailed: %v\npending: %v\n", len(succeeded), len(skipped), failed, pending)
	_ = os.WriteFile(summaryPath, []byte(summary), 0o644)

	if len(pending) > 0 {
		inspect("no CI checks still running",
			false,
			summaryPath,
			strings.Join(pending, "; "))
	}

	inspect("all PR CI checks pass (no failures)",
		len(failed) == 0 && len(succeeded) > 0,
		summaryPath,
		strings.Join(failed, "; "))

	// CHECK 4 (optional): gh pr view when authenticated
	ghOut, ghErr := run("gh", "pr", "view", prNumber, "--json", "state,title,statusCheckRollup")
	ghPath := filepath.Join(outDir(), "gh-pr-view.json")
	_ = os.WriteFile(ghPath, []byte(ghOut), 0o644)
	if ghErr != nil || strings.Contains(ghOut, "HTTP 401") {
		fmt.Printf("CHECK: gh pr view (optional)\n")
		fmt.Printf("EVIDENCE: %s\n", ghPath)
		fmt.Printf("RESULT: SKIP\n")
		fmt.Printf("REASON: gh not authenticated; github-fetch ci used instead\n")
	} else {
		var pr prView
		if err := json.Unmarshal([]byte(ghOut), &pr); err == nil {
			inspect("PR state is OPEN (gh)",
				strings.EqualFold(pr.State, "OPEN"),
				pr.State,
				"PR is not open")
		}
	}

	fmt.Println("ALL CHECKS PASSED — PR #" + prNumber + " ready")
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}