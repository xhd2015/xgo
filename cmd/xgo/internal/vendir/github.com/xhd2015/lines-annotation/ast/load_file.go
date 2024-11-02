package ast

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func LoadFile(dir string, relFile string) (LoadInfo, error) {
	content, err := os.ReadFile(filepath.Join(dir, relFile))
	if err != nil {
		return nil, err
	}
	return loadCode(token.NewFileSet(), relFile, content)
}

func LoadFiles(dir string, relFiles []string) (LoadInfo, error) {
	astFiles := make([]*file, 0, len(relFiles))
	fset := token.NewFileSet()
	for _, relFile := range relFiles {
		ast, err := loadFile(fset, filepath.Join(dir, relFile), relFile)
		if err != nil {
			return nil, err
		}
		astFiles = append(astFiles, ast)
	}

	return &dirLoadInfo{
		dir:   dir,
		fset:  fset,
		files: astFiles,
	}, nil
}

func loadFile(fset *token.FileSet, path string, relPath string) (*file, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ast, syntaxErr := parser.ParseFile(fset, relPath, content, parser.ParseComments)
	if err != nil {
		return nil, nil
	}
	return &file{
		relPath:   relPath,
		ast:       ast,
		content:   content,
		syntaxErr: syntaxErr,
	}, nil
}
