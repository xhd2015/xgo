package instrument_go

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/goinfo"
)

const execStubGo126 = `package work

func (b *Builder) runCover() {
	infiles = append(infiles, sourceFile)
}

func (b *Builder) buildActionID() {
	fmt.Fprintf(h, "compile\n")
}
`

func TestInstrumentExecGo126PatchesRunCoverInfiles(t *testing.T) {
	goroot := t.TempDir()
	execDir := filepath.Join(goroot, "src", "cmd", "go", "internal", "work")
	if err := os.MkdirAll(execDir, 0o755); err != nil {
		t.Fatal(err)
	}
	execPath := filepath.Join(execDir, "exec.go")
	if err := os.WriteFile(execPath, []byte(execStubGo126), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := instrumentExec(goroot, &goinfo.GoVersion{Major: 1, Minor: 26}); err != nil {
		t.Fatalf("instrumentExec go1.26: %v", err)
	}

	out, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(out)
	if !strings.Contains(content, "func (b *Builder) runCover(") {
		t.Fatalf("runCover anchor missing after patch:\n%s", content)
	}
	if !strings.Contains(content, "/*<begin modify_infiles>*/") {
		t.Fatalf("modify_infiles patch marker missing:\n%s", content)
	}
	if !strings.Contains(content, "infiles = append(infiles, __xgo_overlay_source_file)") {
		t.Fatalf("expected runCover infiles overlay patch, got:\n%s", content)
	}
}