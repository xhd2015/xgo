package syntax

import (
	"cmd/compile/internal/syntax"
	"strings"
	"testing"
)

func parseContent(content string) (*syntax.File, error) {
	fbase := syntax.NewFileBase("test.go")
	r := strings.NewReader(content)
	return syntax.Parse(fbase, r, nil /*nil error handler*/, nil /*ignore progma*/, syntax.CheckBranches)
}

func TestCheckContextArg(t *testing.T) {
	file, err := parseContent("package test; func Test(ctx context.Context){}")
	if err != nil {
		t.Fatal(err)
	}
	fn := file.DeclList[0].(*syntax.FuncDecl)
	param0 := fn.Type.ParamList[0].Type
	// t.Logf("file:%v", file)
	if !hasQualifiedName(param0, "context", "Context") {
		t.Fatalf("expect param[0] to be context.Context, actual: %v", param0)
	}
}
