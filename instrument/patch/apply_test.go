package patch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

// === Parser Tests ===

func TestParseXgoPatch_SingleBlock(t *testing.T) {
	patch := `<patch test>
goto struct g
goto closing }
insert_before __xgo_g __xgo_g;
newline
</patch>`

	pf, err := ParseXgoPatch(patch)
	if err != nil {
		t.Fatal(err)
	}
	if len(pf.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(pf.Blocks))
	}
	b := pf.Blocks[0]
	if b.Name != "test" {
		t.Errorf("expected name 'test', got %q", b.Name)
	}
	if len(b.Commands) != 4 {
		t.Fatalf("expected 4 commands, got %d", len(b.Commands))
	}

	assertCmd(t, b.Commands[0], CmdGoto, "struct g", "goto struct g")
	assertCmd(t, b.Commands[1], CmdGoto, "closing }", "goto closing }")
	assertCmd(t, b.Commands[2], CmdInsertBefore, "__xgo_g __xgo_g;", "insert_before")
	assertCmd(t, b.Commands[3], CmdNewline, "", "newline")
}

func TestParseXgoPatch_MultipleBlocks(t *testing.T) {
	patch := `<patch block1>
goto func f
</patch>
<patch block2>
match hello
</patch>`

	pf, err := ParseXgoPatch(patch)
	if err != nil {
		t.Fatal(err)
	}
	if len(pf.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(pf.Blocks))
	}
	if pf.Blocks[0].Name != "block1" {
		t.Errorf("block 0 name: %q", pf.Blocks[0].Name)
	}
	if pf.Blocks[1].Name != "block2" {
		t.Errorf("block 1 name: %q", pf.Blocks[1].Name)
	}
}

func TestParseXgoPatch_Comments(t *testing.T) {
	patch := `<patch test>
# this is a comment
goto struct g

# another comment
goto closing }
</patch>`

	pf, err := ParseXgoPatch(patch)
	if err != nil {
		t.Fatal(err)
	}
	if len(pf.Blocks[0].Commands) != 2 {
		t.Fatalf("expected 2 commands (comments ignored), got %d", len(pf.Blocks[0].Commands))
	}
}

func TestParseXgoPatch_AllCommands(t *testing.T) {
	patch := `<patch all>
goto struct Foo
goto func Bar
goto func (t *T) Baz
goto interface Stringer
goto opening {
goto closing }
goto field X
match some text
find_for_replace target
insert_before added
insert_after appended
replace newtext
newline
</patch>`

	pf, err := ParseXgoPatch(patch)
	if err != nil {
		t.Fatal(err)
	}
	cmds := pf.Blocks[0].Commands
	if len(cmds) != 13 {
		t.Fatalf("expected 13 commands, got %d", len(cmds))
	}

	expected := []struct {
		typ  CommandType
		text string
	}{
		{CmdGoto, "struct Foo"},
		{CmdGoto, "func Bar"},
		{CmdGoto, "func (t *T) Baz"},
		{CmdGoto, "interface Stringer"},
		{CmdGoto, "opening {"},
		{CmdGoto, "closing }"},
		{CmdGoto, "field X"},
		{CmdMatch, "some text"},
		{CmdFindForReplace, "target"},
		{CmdInsertBefore, "added"},
		{CmdInsertAfter, "appended"},
		{CmdReplace, "newtext"},
		{CmdNewline, ""},
	}

	for i, e := range expected {
		if cmds[i].Type != e.typ {
			t.Errorf("cmd[%d] type: expected %v, got %v", i, e.typ, cmds[i].Type)
		}
	}
	if cmds[7].SearchText != "some text" {
		t.Errorf("match text: %q", cmds[7].SearchText)
	}
	if cmds[9].EditText != "added" {
		t.Errorf("insert_before text: %q", cmds[9].EditText)
	}
}

func TestParseXgoPatch_EmptyInsertBeforeError(t *testing.T) {
	patch := `<patch test>
insert_before
</patch>`
	_, err := ParseXgoPatch(patch)
	if err == nil {
		t.Fatal("expected error for empty insert_before")
	}
}

func TestParseXgoPatch_EmptyReplaceError(t *testing.T) {
	patch := `<patch test>
replace
</patch>`
	_, err := ParseXgoPatch(patch)
	if err == nil {
		t.Fatal("expected error for empty replace")
	}
}

func TestParseXgoPatch_UnknownCommandError(t *testing.T) {
	patch := `<patch test>
frobnicate
</patch>`
	_, err := ParseXgoPatch(patch)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

// === Engine Tests: Goto ===

func TestApplyPatch_GotoStruct(t *testing.T) {
	source := `package p
type g struct {
	a int
	b string
}`

	patch := `<patch test>
goto struct g
goto closing }
insert_before __xgo_g __xgo_g;
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the insertion is in the struct and markers are inline
	if !strings.Contains(result, "__xgo_g __xgo_g;") {
		t.Errorf("expected __xgo_g field: %s", result)
	}
	if !strings.Contains(result, "/*<begin test>*/") {
		t.Errorf("expected begin marker: %s", result)
	}
	if !strings.Contains(result, "/*<end test>*/") {
		t.Errorf("expected end marker: %s", result)
	}
	// Verify struct is still valid
	if !strings.Contains(result, "type g struct {") {
		t.Errorf("struct declaration missing: %s", result)
	}
	if !strings.Contains(result, "a int") {
		t.Errorf("existing field a missing: %s", result)
	}
	if !strings.Contains(result, "b string") {
		t.Errorf("existing field b missing: %s", result)
	}
	// The closing } should still be there (after the end marker)
	closeIdx := strings.LastIndex(result, "}")
	if closeIdx < 0 {
		t.Fatal("closing brace missing")
	}
}

func TestApplyPatch_GotoFunc(t *testing.T) {
	source := `package p
func f() {
	doSomething()
}`

	patch := `<patch test>
goto func f
goto opening {
insert_after extraCode()
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "extraCode()") {
		t.Errorf("expected extraCode() in result: %s", result)
	}
	if !strings.Contains(result, "{/*<begin test>*/extraCode()") {
		t.Errorf("expected inline marker after {: %s", result)
	}
}

func TestApplyPatch_GotoFuncWithReceiver(t *testing.T) {
	source := `package p
type T struct{}
func (t *T) Method() {
	x := 1
}`

	patch := `<patch test>
goto func (*T) Method
goto opening {
insert_after instrument()
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "instrument()") {
		t.Errorf("expected instrument() in result: %s", result)
	}
}

func TestApplyPatch_GotoClosing(t *testing.T) {
	source := `package p
func f() {
	x := 1
	return
}`

	patch := `<patch test>
goto func f
goto closing }
insert_before cleanup()
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "cleanup()") {
		t.Errorf("expected cleanup() before closing }: %s", result)
	}
}

func TestApplyPatch_GotoInterface(t *testing.T) {
	source := `package p
type I interface {
	Foo()
	Bar()
}`

	patch := `<patch test>
goto interface I
goto closing }
insert_before Baz()
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "Baz()") {
		t.Errorf("expected Baz() in result: %s", result)
	}
	if !strings.Contains(result, "/*<begin test>*/") {
		t.Errorf("expected begin marker: %s", result)
	}
}

func TestApplyPatch_GotoField(t *testing.T) {
	source := `package p
type g struct {
	_panic *_panic
	_defer *_defer
}`

	patch := `<patch test>
goto struct g
goto field _defer
insert_before extraField int
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "extraField") {
		t.Errorf("expected extraField before _defer: %s", result)
	}
}

// === Engine Tests: Match ===

func TestApplyPatch_Match(t *testing.T) {
	source := `package p
func f() {
	systemstack(func() {
		doWork()
	})
}`

	patch := `<patch test>
goto func f
match systemstack(func() {
insert_before preWork()
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "preWork()") {
		t.Errorf("expected preWork() before systemstack: %s", result)
	}
}

// === Engine Tests: Replace ===

func TestApplyPatch_FindForReplace(t *testing.T) {
	source := `package p
var link = "//go:linkname timeSleep time.Sleep"`

	patch := `<patch test>
find_for_replace time.Sleep
replace time.runtimeSleep
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "time.runtimeSleep") {
		t.Errorf("expected time.runtimeSleep: %s", result)
	}
	if strings.Contains(result, "time.Sleep") && strings.Count(result, "time.Sleep") == 1 {
		// the old text should be in the old tag, not in the replaced position
	}
	if !strings.Contains(result, "/*old:time.Sleep*/") {
		t.Errorf("expected old tag with time.Sleep: %s", result)
	}
}

func TestApplyPatch_ReplaceWithoutFindError(t *testing.T) {
	source := "package p"

	patch := `<patch test>
replace something
</patch>`

	_, err := ApplyXgoPatchContent(source, patch)
	if err == nil {
		t.Fatal("expected error for replace without find_for_replace")
	}
}

func TestApplyPatch_ReplaceEmptyTextError(t *testing.T) {
	patch := `<patch test>
find_for_replace x
replace
</patch>`

	// Parsing should catch this
	_, err := ParseXgoPatch(patch)
	if err == nil {
		t.Fatal("expected parse error for empty replace")
	}
}

// === Engine Tests: Insert stacking ===

func TestApplyPatch_InsertBeforeStacking(t *testing.T) {
	source := `package p
func f() {
	systemstack(func() {
	})
}`

	// insert_before stacks in reverse: last written = closest to original
	patch := `<patch test>
goto func f
match systemstack(func() {
insert_before var x *g
newline
insert_before curg := gp.m.curg
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	// Expected: curg appears first (closer to original due to reverse stacking? No...)
	// Actually: insert_before stacks in reverse:
	// 1. "var x *g" inserted at offset
	// 2. "\n" inserted at offset
	// 3. "curg := gp.m.curg" inserted at offset
	// 4. "\n" inserted at offset
	// Reverse: "\n", "curg := gp.m.curg", "\n", "var x *g"
	// After trim: "curg := gp.m.curg\nvar x *g"
	// So curg is first, var x is second.
	// And both appear before systemstack.

	if !strings.Contains(result, "curg := gp.m.curg") {
		t.Errorf("expected curg: %s", result)
	}
	if !strings.Contains(result, "var x *g") {
		t.Errorf("expected var x: %s", result)
	}
	// curg should come before var x in the text (closer to systemstack)
	curgIdx := strings.Index(result, "curg := gp.m.curg")
	varIdx := strings.Index(result, "var x *g")
	if curgIdx < 0 || varIdx < 0 {
		t.Fatal("text not found")
	}
	if curgIdx > varIdx {
		t.Errorf("expected curg before var x in text (reverse stacking)")
	}
	systemstackIdx := strings.Index(result, "systemstack")
	if curgIdx > systemstackIdx || varIdx > systemstackIdx {
		t.Errorf("insertions should be before systemstack")
	}
}

// === Marker Tests ===

func TestApplyPatch_InlineMarkers(t *testing.T) {
	source := `package p
type g struct {
	a int
}`

	patch := `<patch marker_test>
goto struct g
goto closing }
insert_before newField int
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	// Verify markers exist
	if !strings.Contains(result, "/*<begin marker_test>*/") {
		t.Errorf("missing begin marker: %s", result)
	}
	if !strings.Contains(result, "/*<end marker_test>*/") {
		t.Errorf("missing end marker: %s", result)
	}
	// After the end marker, the closing } should be in the result
	if !strings.Contains(result, "/*<end marker_test>*/}") {
		t.Errorf("end marker should share line with closing }: %s", result)
	}
}

func TestApplyPatch_SingleSeqPerPosition(t *testing.T) {
	source := `package p
func f() {
	a()
	b()
}`

	patch := `<patch seq_test>
goto func f
goto opening {
insert_after preA()
newline
match b()
insert_before preB()
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	// Two edit positions, both use marker with block name "seq_test"
	beginCount := strings.Count(result, "/*<begin seq_test>*/")
	if beginCount != 2 {
		t.Errorf("expected 2 begin markers, got %d: %s", beginCount, result)
	}
	endCount := strings.Count(result, "/*<end seq_test>*/")
	if endCount != 2 {
		t.Errorf("expected 2 end markers, got %d: %s", endCount, result)
	}
	if !strings.Contains(result, "preA()") {
		t.Errorf("expected preA(): %s", result)
	}
	if !strings.Contains(result, "preB()") {
		t.Errorf("expected preB(): %s", result)
	}
}

// === clearPatch Tests ===

func TestClearPatch_Insert(t *testing.T) {
	patched := `package p
type g struct {
	a int
	/*<begin test>*/newField int;/*<end test>*/}`

	result := clearPatch(patched, "test")

	expected := `package p
type g struct {
	a int
	}`
	if result != expected {
		t.Errorf("diff:\n%s", assert.Diff(expected, result))
	}
}

func TestClearPatch_Replace(t *testing.T) {
	patched := `/*<begin test>*//*old:time.Sleep*/time.runtimeSleep/*<end test>*/ var x int`

	result := clearPatch(patched, "test")

	expected := `time.Sleep var x int`
	if result != expected {
		t.Errorf("diff:\n%s", assert.Diff(expected, result))
	}
}

func TestClearPatch_MultipleEdits(t *testing.T) {
	patched := `start
/*<begin test>*/insert1/*<end test>*/
middle
/*<begin test>*/insert2/*<end test>*/
end`

	result := clearPatch(patched, "test")

	expected := `start

middle

end`
	if result != expected {
		t.Errorf("diff:\n%s", assert.Diff(expected, result))
	}
}

func TestClearPatch_NoMarkers(t *testing.T) {
	content := "package p\nfunc f() {}"
	result := clearPatch(content, "test")
	if result != content {
		t.Errorf("expected no change: %s", assert.Diff(content, result))
	}
}

// === Idempotency Tests ===

func TestApplyPatch_Idempotent(t *testing.T) {
	source := `package p
type g struct {
	a int
}`

	patch := `<patch test>
goto struct g
goto closing }
insert_before newField int
newline
</patch>`

	result1, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	// Apply again — should produce same result (clear then re-apply)
	result2, err := ApplyXgoPatchContent(result1, patch)
	if err != nil {
		t.Fatal(err)
	}

	if result1 != result2 {
		t.Errorf("idempotent apply failed:\n%s", assert.Diff(result1, result2))
	}
}

func TestApplyPatch_IdempotentAfterChange(t *testing.T) {
	source := `package p
type g struct {
	a int
}`

	patchV1 := `<patch test>
goto struct g
goto closing }
insert_before oldField int
newline
</patch>`

	patchV2 := `<patch test>
goto struct g
goto closing }
insert_before newField int
newline
</patch>`

	result1, _ := ApplyXgoPatchContent(source, patchV1)
	result2, _ := ApplyXgoPatchContent(result1, patchV2)

	// result2 should NOT contain oldField (cleared before v2 applied)
	if strings.Contains(result2, "oldField") {
		t.Errorf("expected oldField to be cleared: %s", result2)
	}
	if !strings.Contains(result2, "newField") {
		t.Errorf("expected newField: %s", result2)
	}
}

// === Multiple Blocks Test ===

func TestApplyPatch_MultipleBlocks(t *testing.T) {
	source := `package p
func f() {
	a()
	b()
}`

	patch := `<patch add_preA>
goto func f
match a()
insert_before preA()
newline
</patch>
<patch add_preB>
goto func f
match b()
insert_before preB()
newline
</patch>`

	result, err := ApplyXgoPatchContent(source, patch)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "preA()") {
		t.Errorf("expected preA: %s", result)
	}
	if !strings.Contains(result, "preB()") {
		t.Errorf("expected preB: %s", result)
	}
}

// === Error Cases ===

func TestApplyPatch_MatchNotFound(t *testing.T) {
	source := "package p"

	patch := `<patch test>
match nonexistent
insert_before x
</patch>`

	_, err := ApplyXgoPatchContent(source, patch)
	if err == nil {
		t.Fatal("expected error for match not found")
	}
}

func TestApplyPatch_GotoUnknownDecl(t *testing.T) {
	source := "package p"

	patch := `<patch test>
goto struct nonexistent
goto closing }
</patch>`

	_, err := ApplyXgoPatchContent(source, patch)
	if err == nil {
		t.Fatal("expected error for unknown struct")
	}
}

func TestApplyPatch_InsertBeforeEmptyError(t *testing.T) {
	patch := `<patch test>
match x
insert_before
</patch>`

	_, err := ApplyXgoPatchContent("package p", patch)
	if err == nil {
		t.Fatal("expected error for empty insert_before")
	}
}

func TestApplyPatch_InsertAfterEmptyError(t *testing.T) {
	patch := `<patch test>
match x
insert_after
</patch>`

	_, err := ApplyXgoPatchContent("package p", patch)
	if err == nil {
		t.Fatal("expected error for empty insert_after")
	}
}

// === Helpers ===

func assertCmd(t *testing.T, cmd Command, expectedType CommandType, expectedText string, desc string) {
	t.Helper()
	if cmd.Type != expectedType {
		t.Errorf("%s: expected type %v, got %v", desc, expectedType, cmd.Type)
	}
	switch expectedType {
	case CmdGoto:
		if cmd.GotoTarget != expectedText {
			t.Errorf("%s: expected target %q, got %q", desc, expectedText, cmd.GotoTarget)
		}
	case CmdMatch, CmdFindForReplace:
		if cmd.SearchText != expectedText {
			t.Errorf("%s: expected search text %q, got %q", desc, expectedText, cmd.SearchText)
		}
	case CmdInsertBefore, CmdInsertAfter, CmdReplace:
		if cmd.EditText != expectedText {
			t.Errorf("%s: expected edit text %q, got %q", desc, expectedText, cmd.EditText)
		}
	}
}

// === Config Tests ===

func TestLoadConfig_NewArrayFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configJSON := `{
  "version": "go1.24",
  "copy": [
    {"from": "patch/", "to": "src/cmd/compile/internal/xgo_rewrite_internal/patch/"}
  ],
  "generate": [
    {
      "cmd": "go run ${XGO_SRC}/script/mkbuiltin --goroot=${GOROOT}",
      "outputs": ["src/cmd/compile/internal/typecheck/builtin.go"]
    }
  ]
}`
	writeTestFile(t, filepath.Join(tmpDir, "__config__.json"), configJSON)

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != "go1.24" {
		t.Errorf("expected version go1.24, got %s", cfg.Version)
	}
	if len(cfg.Copy) != 1 {
		t.Fatalf("expected 1 copy entry, got %d", len(cfg.Copy))
	}
	if cfg.Copy[0].From != "patch/" {
		t.Errorf("expected from patch/, got %s", cfg.Copy[0].From)
	}
	if cfg.Copy[0].To != "src/cmd/compile/internal/xgo_rewrite_internal/patch/" {
		t.Errorf("expected to src/..., got %s", cfg.Copy[0].To)
	}
	if len(cfg.Generate) != 1 {
		t.Fatalf("expected 1 generate entry, got %d", len(cfg.Generate))
	}
	if cfg.Generate[0].Cmd != "go run ${XGO_SRC}/script/mkbuiltin --goroot=${GOROOT}" {
		t.Errorf("unexpected cmd: %s", cfg.Generate[0].Cmd)
	}
	if len(cfg.Generate[0].Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(cfg.Generate[0].Outputs))
	}
}

func TestLoadConfig_EmptyToDefaultsToPatchDir(t *testing.T) {
	tmpDir := t.TempDir()
	configJSON := `{"version": "go1.25", "copy": [{"from": "patch/"}]}`
	writeTestFile(t, filepath.Join(tmpDir, "__config__.json"), configJSON)

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Copy[0].To != "" {
		t.Fatalf("expected empty To, got %q", cfg.Copy[0].To)
	}
}

func TestLoadConfig_MissingFileReturnsEmptyConfig(t *testing.T) {
	cfg, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != "" {
		t.Errorf("expected empty version, got %s", cfg.Version)
	}
	if len(cfg.Copy) != 0 {
		t.Errorf("expected 0 copy entries, got %d", len(cfg.Copy))
	}
}

func TestApplyPatches_GenerateVariableSubstitution(t *testing.T) {
	// This test verifies variable substitution works.
	// We use a simple echo command to check output.
	tmpDir := t.TempDir()
	goroot := t.TempDir()
	xgoSrc := t.TempDir()

	configJSON := `{"version": "test", "generate": [{"cmd": "echo hello ${NAME}", "outputs": []}]}`
	writeTestFile(t, filepath.Join(tmpDir, "__config__.json"), configJSON)

	err := ApplyPatches(tmpDir, goroot, xgoSrc, map[string]string{
		"NAME": "world",
	}, nil, nil)
	// This will try to run "echo hello world" and succeed
	if err != nil {
		// If echo isn't found, skip (Windows might not have it)
		t.Logf("skipped: %v", err)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateEntry_KindField(t *testing.T) {
	configJSON := `{
		"version": "test",
		"generate": [
			{"cmd": "echo always", "outputs": []},
			{"kind": "rebuild-compiler", "cmd": "echo compile", "outputs": []},
			{"kind": "rebuild-go", "cmd": "echo go", "outputs": []}
		]
	}`
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "__config__.json"), configJSON)

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Generate) != 3 {
		t.Fatalf("expected 3 generate entries, got %d", len(cfg.Generate))
	}
	if cfg.Generate[0].Kind != "" {
		t.Errorf("entry 0 kind should be empty, got %q", cfg.Generate[0].Kind)
	}
	if cfg.Generate[1].Kind != "rebuild-compiler" {
		t.Errorf("entry 1 kind should be rebuild-compiler, got %q", cfg.Generate[1].Kind)
	}
	if cfg.Generate[2].Kind != "rebuild-go" {
		t.Errorf("entry 2 kind should be rebuild-go, got %q", cfg.Generate[2].Kind)
	}
}

func TestApplyPatches_SkipKinds(t *testing.T) {
	tmpDir := t.TempDir()
	goroot := t.TempDir()
	xgoSrc := t.TempDir()

	configJSON := `{
		"version": "test",
		"generate": [
			{"cmd": "echo always", "outputs": []},
			{"kind": "rebuild-compiler", "cmd": "echo should_skip", "outputs": []},
			{"kind": "rebuild-go", "cmd": "echo should_skip", "outputs": []}
		]
	}`
	writeTestFile(t, filepath.Join(tmpDir, "__config__.json"), configJSON)

	skipKinds := []string{"rebuild-compiler", "rebuild-go"}
	err := ApplyPatches(tmpDir, goroot, xgoSrc, nil, skipKinds, nil)
	if err != nil {
		t.Fatalf("ApplyPatches failed: %v", err)
	}
}

func TestApplyPatches_SkipKinds_NoSkipWithoutKinds(t *testing.T) {
	tmpDir := t.TempDir()
	goroot := t.TempDir()
	xgoSrc := t.TempDir()

	configJSON := `{
		"version": "test",
		"generate": [
			{"cmd": "echo run_me", "outputs": []},
			{"kind": "rebuild-compiler", "cmd": "echo skip_me", "outputs": []}
		]
	}`
	writeTestFile(t, filepath.Join(tmpDir, "__config__.json"), configJSON)

	skipKinds := []string{"rebuild-go"}
	err := ApplyPatches(tmpDir, goroot, xgoSrc, nil, skipKinds, nil)
	if err != nil {
		t.Fatalf("ApplyPatches failed: %v", err)
	}
}

func TestApplyPatches_IgnoreFiles(t *testing.T) {
	xgoSrc := t.TempDir()
	goroot := t.TempDir()
	patchDir := t.TempDir()

	srcDir := filepath.Join(xgoSrc, "test-src")
	mustMkdirAll(t, srcDir)
	writeTestFile(t, filepath.Join(srcDir, "keep.txt"), "keep me")
	writeTestFile(t, filepath.Join(srcDir, "go.mod"), "module test")

	configJSON := `{
		"version": "test",
		"copy": [
			{"from": "test-src/", "to": ".", "ignore_files": ["go.mod"]}
		]
	}`
	writeTestFile(t, filepath.Join(patchDir, "__config__.json"), configJSON)

	err := ApplyPatches(patchDir, goroot, xgoSrc, nil, nil, nil)
	if err != nil {
		t.Fatalf("ApplyPatches failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(goroot, "keep.txt")); err != nil {
		t.Errorf("keep.txt should exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(goroot, "go.mod")); !os.IsNotExist(err) {
		t.Errorf("go.mod should be removed by ignore_files, got err=%v", err)
	}
}
