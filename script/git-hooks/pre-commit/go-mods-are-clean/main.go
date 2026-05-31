package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	rootDir, err := getRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-mods-are-clean: %v\n", err)
		os.Exit(1)
	}

	modFiles := []string{
		filepath.Join(rootDir, "go.mod"),
		filepath.Join(rootDir, "runtime", "go.mod"),
	}

	hasError := false
	for _, modFile := range modFiles {
		if err := checkGoMod(modFile); err != nil {
			fmt.Fprintf(os.Stderr, "go-mods-are-clean: %v\n", err)
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
	}
}

func getRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to find git root: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

var dirtyDirectives = map[string]bool{
	"require": true,
	"replace": true,
	"exclude": true,
	"retract": true,
}

func checkGoMod(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return checkGoModContent(filepath.Base(filePath), string(data))
}

func checkGoModContent(fileName string, content string) error {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		token := strings.TrimSpace(strings.SplitN(trimmed, " ", 2)[0])
		if dirtyDirectives[token] {
			return fmt.Errorf("%s must have no dependencies: found %q directive in %s", fileName, token, fileName)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
