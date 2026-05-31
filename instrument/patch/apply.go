package patch

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/support/filecopy"
)

// CopyEntry represents a directory copy instruction in __config__.json.
type CopyEntry struct {
	From        string   `json:"from"` // source relative to xgo repo root
	To          string   `json:"to"`   // destination relative to GOROOT; empty = patchDir
	IgnoreFiles []string `json:"ignore_files,omitempty"`
}

// GenerateEntry represents a shell command to run during patching.
type GenerateEntry struct {
	Kind    string   `json:"kind,omitempty"`
	Cmd     string   `json:"cmd"`
	Outputs []string `json:"outputs"`
}

// Config represents the __config__.json file in a patch directory.
type Config struct {
	Version  string          `json:"version"`
	Copy     []CopyEntry     `json:"copy,omitempty"`
	Generate []GenerateEntry `json:"generate,omitempty"`
}

// ApplyPatches walks a patch directory and applies all operations to a GOROOT.
// Auto-discovers:
//   - .xgo.patch files → applied to corresponding GOROOT files
//   - Other files (except __config__.json) → copied one-to-one to GOROOT
//   - __config__.json → copy-dir instructions + generate commands
//
// extraEnv provides variable substitution for ${VAR} in generate commands.
// skipKinds, if non-empty, skips generate entries whose Kind field matches any value.
func ApplyPatches(patchDir, goroot, xgoRepoRoot string, extraEnv map[string]string, skipKinds []string) error {
	cfg, err := LoadConfig(patchDir)
	if err != nil {
		return fmt.Errorf("load __config__.json: %w", err)
	}

	for _, entry := range cfg.Copy {
		srcPath := filepath.Join(xgoRepoRoot, entry.From)
		dstPath := entry.To
		if dstPath == "" {
			dstPath = "." + string(filepath.Separator)
		}
		dstPath = filepath.Join(goroot, dstPath)
		opts := filecopy.NewOptions()
		if config.IS_DEV {
			err = opts.UseLink().CopyReplaceDir(srcPath, dstPath)
		} else {
			err = opts.Copy(srcPath, dstPath)
		}
		if err != nil {
			return fmt.Errorf("copy dir %s -> %s: %w", entry.From, dstPath, err)
		}
		for _, ignoreFile := range entry.IgnoreFiles {
			rmPath := filepath.Join(goroot, ignoreFile)
			if err := os.RemoveAll(rmPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("remove ignored file %s: %w", rmPath, err)
			}
		}
	}

	for _, gen := range cfg.Generate {
		if shouldSkipKind(gen.Kind, skipKinds) {
			continue
		}
		cmdStr := gen.Cmd
		for k, v := range extraEnv {
			cmdStr = strings.ReplaceAll(cmdStr, "${"+k+"}", v)
		}
		// simple shell execution: split on spaces for command and args
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}
		execCmd := exec.Command(parts[0], parts[1:]...)
		execCmd.Dir = goroot
		execCmd.Env = os.Environ()
		for i, e := range execCmd.Env {
			if strings.HasPrefix(e, "GOROOT=") {
				execCmd.Env = append(execCmd.Env[:i], execCmd.Env[i+1:]...)
				break
			}
		}
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("generate %q: %w", gen.Cmd, err)
		}
	}

	return filepath.Walk(patchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == "__config__.json" {
			return nil
		}

		relPath, err := filepath.Rel(patchDir, path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(relPath, ".xgo.patch") {
			targetRel := strings.TrimSuffix(relPath, ".xgo.patch")
			targetPath := filepath.Join(goroot, targetRel)
			return ApplyXgoPatchFile(path, targetPath)
		}

		targetPath := filepath.Join(goroot, relPath)
		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return filecopy.CopyFileAll(path, targetPath)
	})
}

// ApplyXgoPatchFile reads a .xgo.patch file and applies it to the target file.
// The target file is modified in place; existing edits with the same patch name
// are cleared first (idempotent).
func ApplyXgoPatchFile(patchFile, targetFile string) error {
	patchContent, err := os.ReadFile(patchFile)
	if err != nil {
		return err
	}

	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}

	result, err := ApplyXgoPatchContent(string(targetContent), string(patchContent))
	if err != nil {
		return fmt.Errorf("%s: %w", patchFile, err)
	}

	return os.WriteFile(targetFile, []byte(result), 0644)
}

// ApplyXgoPatchContent applies parsed patch content to a source string.
// Each <patch> block is applied independently; blocks re-parse after prior block edits.
// Before applying a named patch, any existing markers with that name are cleared.
func ApplyXgoPatchContent(source string, patchContent string) (string, error) {
	pf, err := ParseXgoPatch(patchContent)
	if err != nil {
		return "", err
	}

	result := source
	for _, block := range pf.Blocks {
		if block.Name != "" {
			result = clearPatch(result, block.Name)
		}

		var err error
		result, err = applyPatch(result, block)
		if err != nil {
			return "", fmt.Errorf("patch %q: %w", block.Name, err)
		}
	}

	return result, nil
}

// clearPatch removes all marker content for a named patch from the source text.
// For replace edits, the /*<old:...>*/ content is restored.
// For replace_directive, the /*<next-line-original> annotation is used to restore.
func clearPatch(content string, patchName string) string {
	// First pass: handle <next-line-original> annotations
	annotStart := "/*<next-line-original " + patchName + ">"
	annotEnd := "</next-line-original>*/"
	for {
		annotIdx := strings.Index(content, annotStart)
		if annotIdx < 0 {
			break
		}
		annotContentStart := annotIdx + len(annotStart)
		annotCloseIdx := strings.Index(content[annotContentStart:], annotEnd)
		if annotCloseIdx < 0 {
			break
		}
		annotCloseIdx += annotContentStart
		oldText := content[annotContentStart:annotCloseIdx]
		nextLineStart := annotCloseIdx + len(annotEnd)
		var nextLineEnd int
		firstNL := strings.Index(content[nextLineStart:], "\n")
		if firstNL < 0 {
			nextLineEnd = len(content)
		} else {
			afterFirst := nextLineStart + firstNL + 1
			secondNL := strings.Index(content[afterFirst:], "\n")
			if secondNL < 0 {
				nextLineEnd = len(content)
			} else {
				nextLineEnd = afterFirst + secondNL
			}
		}
		content = content[:annotIdx] + oldText + content[nextLineEnd:]
	}

	beginMarker := "/*<begin " + patchName + ">"
	endMarker := "/*<end " + patchName + ">*/"

	for {
		beginIdx := strings.Index(content, beginMarker)
		if beginIdx < 0 {
			break
		}

		// Find end of begin marker (the */ that closes the start comment)
		beginEnd := strings.Index(content[beginIdx:], "*/")
		if beginEnd < 0 {
			break
		}
		beginEnd += beginIdx + 2

		// Find corresponding end marker
		endIdx := strings.Index(content[beginEnd:], endMarker)
		if endIdx < 0 {
			break
		}
		endIdx += beginEnd
		endEnd := endIdx + len(endMarker)

		// Check for old content as /*old:...*/ comment between begin and end markers
		contentBetween := content[beginEnd:endIdx]
		oldText := extractOldContent(contentBetween)
		if oldText != "" {
			// Replace the entire marker block with the old text
			content = content[:beginIdx] + oldText + content[endEnd:]
			continue
		}

		// Remove the marker block (insert content)
		content = content[:beginIdx] + content[endEnd:]
	}

	return content
}

func shouldSkipKind(kind string, skipKinds []string) bool {
	if kind == "" || len(skipKinds) == 0 {
		return false
	}
	for _, skip := range skipKinds {
		if skip == kind {
			return true
		}
	}
	return false
}

// LoadConfig reads and parses the __config__.json file from a patch directory.
// Returns an empty Config if the file does not exist.
func LoadConfig(patchDir string) (*Config, error) {
	configPath := filepath.Join(patchDir, "__config__.json")
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse __config__.json: %w", err)
	}
	return &cfg, nil
}
