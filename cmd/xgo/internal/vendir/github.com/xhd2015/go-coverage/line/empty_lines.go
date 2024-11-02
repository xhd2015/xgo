package line

import (
	"context"
	goast "go/ast"
	"go/token"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/golang.org/x/tools/go/packages"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/ast"
)

func CollectEmptyLines(ctx context.Context, dir string, args []string, buildFlags []string) (map[string][]int, error) {
	absDir, fset, pkgs, err := ast.LoadSyntaxOnly(ctx, dir, BuildArgsToSyntaxArgs(args), buildFlags)
	if err != nil {
		return nil, err
	}
	emptyLinesMapping := make(map[string][]int)
	packages.Visit(pkgs, func(p *packages.Package) bool {
		for _, f := range p.Syntax {
			tokenFile := fset.File(f.Package)
			fileName := tokenFile.Name()
			if !strings.HasPrefix(fileName, absDir) {
				continue
			}
			emptyLinesMapping[strings.TrimPrefix(fileName[len(absDir):], "/")] = CollectEmptyLinesForFile(fset, f)
		}
		return true
	}, nil)
	return emptyLinesMapping, nil
}

// TODO: export
func BuildArgsToSyntaxArgs(args []string) []string {
	if len(args) == 0 {
		return []string{"./..."}
	}
	newArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasPrefix(arg, "./") || strings.HasSuffix(arg, "...") {
			newArgs = append(newArgs, arg)
			continue
		}
		newArgs = append(newArgs, strings.TrimSuffix(arg, "/")+"/...")
	}
	return newArgs
}

// NOTE: only collect lines between ast.File, which is the only possible semantic area.
func CollectEmptyLinesForFile(fset *token.FileSet, f *goast.File) []int {
	maxLine := -1

	//
	lineHavingNonEmptyNode := make(map[int]bool)
	goast.Inspect(f, func(n goast.Node) bool {
		if n == nil {
			// post
			return false // when post, return value is not used
		}
		line := fset.Position(n.Pos()).Line
		endLine := -1
		// the end position example:
		//   1. {
		//	 2.
		//   3. }X
		//  will always remain valid for any node, and will
		//  almostly remain the same line with col+1,because there is a \n.
		if n.End().IsValid() {
			endLine = fset.Position(n.End()).Line
		}
		if maxLine < line {
			maxLine = line
		}
		if endLine != -1 && maxLine < endLine {
			maxLine = endLine
		}

		// check if comment
		switch n.(type) {
		case *goast.Comment, *goast.CommentGroup:
		default:
			lineHavingNonEmptyNode[line] = true
			if n.End().IsValid() {
				lineHavingNonEmptyNode[endLine] = true
			}
		}
		return true
	})

	var emptyLines []int
	for line := 1; line <= maxLine; line++ {
		if !lineHavingNonEmptyNode[line] {
			emptyLines = append(emptyLines, line)
		}
	}

	return emptyLines
}
