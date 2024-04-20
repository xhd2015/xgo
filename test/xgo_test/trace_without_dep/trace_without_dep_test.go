package trace_without_dep

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTraceWithoutDep(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	projectRoot := wd
	for i := 0; i < 3; i++ {
		projectRoot = filepath.Dir(projectRoot)
	}

	targetDir := filepath.Join(wd, "target")
	traceFile := filepath.Join(targetDir, "TestGreet.json")
	os.RemoveAll(traceFile)

	var cmdBuf bytes.Buffer
	cmd := exec.Command(
		"go", "run", "-tags", "dev", "./cmd/xgo",
		"test",
		"--strace",
		"-count=1",
		// "--log-debug=stdout",
		"--project-dir", filepath.Join(wd, "target"),
		"-v",
		"./",
	)
	cmd.Dir = projectRoot
	cmd.Stderr = &cmdBuf
	cmd.Stdout = &cmdBuf
	err = cmd.Run()
	if err != nil {
		t.Fatalf("target test: %s %v", cmdBuf.String(), err)
	}
	_, err = os.ReadFile(traceFile)
	if err != nil {
		t.Fatalf("trace file not generated: %v", err)
	}
}
