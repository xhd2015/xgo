package ast

import (
	"go/parser"
	"go/token"
)

type singleFileLoadInfo struct {
	fileSet *token.FileSet
	ast     *file
}

func (c *singleFileLoadInfo) FileSet() *token.FileSet {
	return c.fileSet
}

// RangeFiles implements AstLoadInfo.
func (s *singleFileLoadInfo) RangeFiles(handler func(f File) bool) {
	handler(s.ast)
}

func LoadCode(relPath string, code []byte) (LoadInfo, error) {
	return loadCode(token.NewFileSet(), relPath, code)
}

func loadCode(fset *token.FileSet, relPath string, code []byte) (LoadInfo, error) {
	ast, err := parseCode(relPath, code, fset)
	if err != nil {
		return nil, err
	}

	return &singleFileLoadInfo{
		fileSet: fset,
		ast:     ast,
	}, nil
}

func parseCode(relPath string, code []byte, fset *token.FileSet) (*file, error) {
	ast, err := parser.ParseFile(fset, relPath, code, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return &file{
		relPath: relPath,
		ast:     ast,
		content: code,
	}, nil
}
