package patch

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/xhd2015/xgo/support/fileutil"
	"github.com/xhd2015/xgo/support/goparse"
	"github.com/xhd2015/xgo/support/transform/astdiff"
	"github.com/xhd2015/xgo/support/transform/edit/line"
	"github.com/xhd2015/xgo/support/transform/patch/unpatch"
)

// see https://github.com/xhd2015/xgo/issues/169#issuecomment-2241407305
func Patch(old string, patch string) (string, error) {
	old, err := unpatch.Unpatch(old)
	if err != nil {
		return "", err
	}
	srcAST, srcFset, err := goparse.ParseFileCode("old.go", []byte(old))
	if err != nil {
		return "", err
	}

	patchCodeAST, patchFset, err := goparse.ParseFileCode("patch.go", []byte(patch))
	if err != nil {
		return "", err
	}
	return patchAST(old, srcAST, srcFset, patch, patchCodeAST, patchFset)
}

func PatchFile(srcFile string, patchFile string) (string, error) {
	srcCode, err := fileutil.ReadFile(srcFile)
	if err != nil {
		return "", err
	}

	unpatchedSrcCode, err := unpatch.Unpatch(string(srcCode))
	if err != nil {
		return "", err
	}

	srcAST, srcFset, err := goparse.ParseFileCode(srcFile, []byte(unpatchedSrcCode))
	if err != nil {
		return "", err
	}

	patchCode, patchCodeAST, patchFset, err := goparse.Parse(patchFile)
	if err != nil {
		return "", err
	}

	return patchAST(unpatchedSrcCode, srcAST, srcFset, string(patchCode), patchCodeAST, patchFset)
}

func patchAST(srcCode string, srcAST *ast.File, srcFset *token.FileSet, patchCode string, patchAST *ast.File, patchFset *token.FileSet) (string, error) {
	patchLines := strings.Split(patchCode, "\n")

	patchMapping := parsePatch(patchLines, patchAST, patchFset)

	var srcSpecs []ast.Spec
	var patchSpecs []ast.Spec

	var srcFuncDecls []*ast.FuncDecl
	var patchFuncDecls []*ast.FuncDecl

	var srcImports []*ast.ImportSpec
	var patchImports []*ast.ImportSpec

	patchFuncMapping := make(map[string]*ast.FuncDecl)
	for _, decl := range patchAST.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			patchFuncMapping[decl.Name.Name] = decl
			patchFuncDecls = append(patchFuncDecls, decl)
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ImportSpec:
					patchImports = append(patchImports, spec)
				default:
					patchSpecs = append(patchSpecs, spec)
				}
			}
		default:
			return "", fmt.Errorf("unhandled %T", decl)
		}
	}

	for _, decl := range srcAST.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			srcFuncDecls = append(srcFuncDecls, decl)
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ImportSpec:
					srcImports = append(srcImports, spec)
				default:
					srcSpecs = append(srcSpecs, spec)
				}
			}
		}
	}
	edit := &line.Edit{}
	// imports
	patchNodeStrict(edit, srcFset, patchMapping, importSpecs(srcImports), importSpecs(patchImports))

	// specs
	patchNodeStrict(edit, srcFset, patchMapping, specs(srcSpecs), specs(patchSpecs))

	// decls
	patchNodes(edit, srcFset, patchMapping, funcDecls(srcFuncDecls), funcDecls(patchFuncDecls), true)

	// inside funcs
	for _, decl := range srcAST.Decls {
		fdecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		fnPatch := patchFuncMapping[fdecl.Name.Name]
		if fnPatch == nil {
			continue
		}

		// check name and type
		if !astdiff.FuncDeclSameIgnoreBody(fdecl, fnPatch) {
			return "", fmt.Errorf("func not match %s: %s -> %s", fdecl.Name.Name, getText(srcCode, fdecl.Type, srcFset), getText(patchCode, fnPatch.Type, patchFset))
		}
		patchFunc(edit, fdecl, srcFset, fnPatch, patchMapping)
	}

	srcLines := strings.Split(srcCode, "\n")
	newLines, err := edit.Apply(srcLines)
	if err != nil {
		return "", err
	}
	newCode := strings.Join(newLines, "\n")
	return newCode, nil
}

type nodeList interface {
	Len() int
	Index(i int) ast.Node
	Slice(i int, j int) nodeList
}

type stmtList []ast.Stmt

func (c stmtList) Len() int {
	return len(c)
}
func (c stmtList) Index(i int) ast.Node {
	stmt := c[i]
	if stmt == nil {
		return nil
	}
	return stmt
}
func (c stmtList) Slice(i int, j int) nodeList {
	if j == -1 {
		j = len(c)
	}
	return stmtList(c[i:j])
}

type importSpecs []*ast.ImportSpec

func (c importSpecs) Len() int {
	return len(c)
}
func (c importSpecs) Index(i int) ast.Node {
	spec := c[i]
	if spec == nil {
		return nil
	}
	return spec
}
func (c importSpecs) Slice(i int, j int) nodeList {
	if j == -1 {
		j = len(c)
	}
	return importSpecs(c[i:j])
}

type specs []ast.Spec

func (c specs) Len() int {
	return len(c)
}
func (c specs) Index(i int) ast.Node {
	spec := c[i]
	if spec == nil {
		return nil
	}
	return spec
}
func (c specs) Slice(i int, j int) nodeList {
	if j == -1 {
		j = len(c)
	}
	return specs(c[i:j])
}

type funcDecls []*ast.FuncDecl

func (c funcDecls) Len() int {
	return len(c)
}
func (c funcDecls) Index(i int) ast.Node {
	spec := c[i]
	if spec == nil {
		return nil
	}
	return spec
}
func (c funcDecls) Slice(i int, j int) nodeList {
	if j == -1 {
		j = len(c)
	}
	return funcDecls(c[i:j])
}

func patchFunc(edit *line.Edit, srcFunc *ast.FuncDecl, srcFset *token.FileSet, patchFunc *ast.FuncDecl, patchMapping map[ast.Node]*PatchContent) {
	var srcStmts []ast.Stmt
	var patchStmts []ast.Stmt

	if srcFunc.Body != nil {
		srcStmts = srcFunc.Body.List
	}
	if patchFunc.Body != nil {
		patchStmts = patchFunc.Body.List
	}

	patchNodeStrict(edit, srcFset, patchMapping, stmtList(srcStmts), stmtList(patchStmts))

}
func patchNodeStrict(edit *line.Edit, srcFset *token.FileSet, patchMapping map[ast.Node]*PatchContent, srcNodes nodeList, patchNodeList nodeList) {
	patchNodes(edit, srcFset, patchMapping, srcNodes, patchNodeList, false)
}

func patchNodes(edit *line.Edit, srcFset *token.FileSet, patchMapping map[ast.Node]*PatchContent, srcNodes nodeList, patchNodes nodeList, funcDeclIgnoreBody bool) {
	var srcIndex int
	var patchIndex int
	for {
		if patchIndex >= patchNodes.Len() {
			break
		}
		if srcIndex >= srcNodes.Len() {
			panic(fmt.Errorf("bad patch"))
		}
		patchHead := patchNodes.Index(patchIndex)
		m := findMatch(srcNodes.Slice(srcIndex, -1), patchHead, funcDeclIgnoreBody)
		if m < 0 {
			panic(fmt.Errorf("patch not found: %v", patchHead))
		}
		srcNode := srcNodes.Index(srcIndex + m)
		srcIndex += m
		patchIndex++
		patch := patchMapping[patchHead]
		if patch == nil {
			continue
		}

		line := srcFset.Position(srcNode.Pos()).Line
		if patch.Append != nil && len(patch.Append.Lines) > 0 {
			edit.Append(line, patch.Append.ID, patch.Append.Lines)
		}
		if patch.Prepend != nil && len(patch.Prepend.Lines) > 0 {
			edit.Prepend(line, patch.Prepend.ID, patch.Prepend.Lines)
		}
		if patch.Replace != nil && len(patch.Replace.Lines) > 0 {
			edit.Replace(line, patch.Replace.ID, patch.Replace.Lines)
		}
	}
}

func getText(code string, node ast.Node, fset *token.FileSet) string {
	start := fset.Position(node.Pos()).Offset
	end := fset.Position(node.End()).Offset
	return code[start:end]
}

func findMatch(list nodeList, node ast.Node, funcDeclIgnoreBody bool) int {
	n := list.Len()
	for i := 0; i < n; i++ {
		srcNode := list.Index(i)
		if funcDeclIgnoreBody {
			srcFuncDecl, ok := srcNode.(*ast.FuncDecl)
			if ok {
				patchFuncDecl, ok := node.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if astdiff.FuncDeclSameIgnoreBody(srcFuncDecl, patchFuncDecl) {
					return i
				}
				continue
			}
		}
		if astdiff.NodeSame(srcNode, node) {
			return i
		}
	}
	return -1
}
