package resolve

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/edit"
)

func Traverse(packages *edit.Packages, recorder *Recorder) error {
	// first, collect all toplevel variables
	for _, pkg := range packages.Packages {
		decls := make(map[string]*edit.Decl)
		for _, file := range pkg.Files {
			var fileDecls []*edit.Decl
			for _, decl := range file.File.Syntax.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					switch decl.Tok {
					case token.VAR:
						for _, spec := range decl.Specs {
							switch spec := spec.(type) {
							case *ast.ValueSpec:
								for _, name := range spec.Names {
									if isBlankName(name.Name) {
										continue
									}
									fileDecls = append(fileDecls, &edit.Decl{
										Kind:  edit.DeclKindVar,
										Ident: name,
										Type:  spec.Type,
										Decl:  decl,
									})
								}
							}
						}
					case token.CONST:
						// TODO
					}
				}
			}
			for _, decl := range fileDecls {
				decls[decl.Ident.Name] = decl
			}
			file.Decls = fileDecls
		}
		pkg.Decls = decls
	}

	// rewrite selectors
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			traverseFuncDecls(packages, pkg, file, recorder)
		}
	}

	return nil
}

func traverseFuncDecls(packages *edit.Packages, pkg *edit.Package, file *edit.File, recorder *Recorder) {
	imports := getFileImports(file.File.Syntax)
	fileScope := newFileScope(&GlobalScope{
		Packages: packages,
		Package:  pkg,
		File:     file,

		PkgScopeNames: pkg.Decls,
		Imports:       imports,
		Recorder:      recorder,
		detectVarTrap: true,
		detectMock:    true,
	})
	// NOTE: statements like this will be ignored:
	//  var _ = func() bool {
	//     x = true()
	//    return true
	// }
	for _, decl := range file.File.Syntax.Decls {
		fnDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fnDecl.Body == nil {
			continue
		}
		fileScope.traverseFunc(fnDecl.Recv, fnDecl.Type, fnDecl.Body)
	}
}
