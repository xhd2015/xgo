package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const prNumber = "389"

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

	// CHECK 3: gh pr view returns status checks
	ghOut, ghErr := run("gh", "pr", "view", prNumber, "--json", "state,title,statusCheckRollup")
	ghPath := filepath.Join(outDir(), "gh-pr-view.json")
	_ = os.WriteFile(ghPath, []byte(ghOut), 0o644)
	inspect("gh pr view returns status checks",
		ghErr == nil && strings.Contains(ghOut, "statusCheckRollup"),
		ghPath,
		strings.TrimSpace(ghOut))

	var pr prView
	if err := json.Unmarshal([]byte(ghOut), &pr); err != nil {
		inspect("parse gh pr view JSON", false, ghPath, err.Error())
	}

	inspect("PR state is OPEN",
		strings.EqualFold(pr.State, "OPEN"),
		pr.State,
		"PR is not open")

	// CHECK 4: all completed checks are success or skipped (none failed)
	var failed []string
	var pending []string
	for _, c := range pr.Status {
		switch strings.ToUpper(c.Conclusion) {
		case "SUCCESS", "SKIPPED", "NEUTRAL":
			// ok
		case "FAILURE", "CANCELLED", "TIMED_OUT", "ACTION_REQUIRED":
			failed = append(failed, c.Name+"="+c.Conclusion)
		case "":
			if strings.ToUpper(c.Status) == "IN_PROGRESS" || strings.ToUpper(c.Status) == "QUEUED" {
				pending = append(pending, c.Name+"="+c.Status)
			}
		default:
			if c.Conclusion != "" {
				failed = append(failed, c.Name+"="+c.Conclusion)
			}
		}
	}

	summaryPath := filepath.Join(outDir(), "check-summary.txt")
	summary := fmt.Sprintf("failed: %v\npending: %v\n", failed, pending)
	_ = os.WriteFile(summaryPath, []byte(summary), 0o644)

	if len(pending) > 0 {
		inspect("no CI checks still running",
			false,
			summaryPath,
			strings.Join(pending, ", "))
	}

	inspect("all PR CI checks pass (no failures)",
		len(failed) == 0,
		summaryPath,
		strings.Join(failed, ", "))

	fmt.Println("ALL CHECKS PASSED — PR #" + prNumber + " ready")
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}