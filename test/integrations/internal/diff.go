package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CompareDirs(dirA, dirB string) string {
	stripA := stripMarkers(dirA)
	defer os.RemoveAll(stripA)
	stripB := stripMarkers(dirB)
	defer os.RemoveAll(stripB)

	baseA := filepath.Base(dirA)
	baseB := filepath.Base(dirB)
	cmd := exec.Command("diff", "-rq",
		"--exclude=.DS_Store",
		"--exclude=bin",
		"--exclude=pkg",
		"--exclude=.git",
		filepath.Join(stripA, baseA, "src"),
		filepath.Join(stripB, baseB, "src"))
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
}

func stripMarkers(dir string) string {
	dst, _ := os.MkdirTemp("", "xgo-compare-*")
	exec.Command("cp", "-R", dir, dst).Run()

	walkRoot := filepath.Join(dst, filepath.Base(dir))
	filepath.Walk(walkRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(content)
		for {
			start := strings.Index(text, "/*<")
			if start < 0 {
				break
			}
			end := strings.Index(text[start:], "*/")
			if end < 0 {
				break
			}
			end += start + 2
			text = text[:start] + text[end:]
		}
		os.WriteFile(path, []byte(text), 0644)
		return nil
	})
	return dst
}
