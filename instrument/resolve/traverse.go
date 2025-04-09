package resolve

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/edit"
)

type Options struct {
	RecordVariables bool
	RecordTypes     bool
}

func Collect(packages *edit.Packages, opts Options) {
	// first, collect all toplevel variables
	for _, pkg := range packages.Packages {
		decls := pkg.Decls
		if decls == nil {
			decls = make(map[string]*edit.Decl)
		}
		for _, file := range pkg.Files {
			fileDecls := file.Decls
			for _, decl := range file.File.Syntax.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					switch decl.Tok {
					case token.VAR:
						for _, spec := range decl.Specs {
							switch spec := spec.(type) {
							case *ast.ValueSpec:
								if opts.RecordVariables && !pkg.HasVarDecls {
									for _, name := range spec.Names {
										if isBlankName(name.Name) {
											continue
										}
										fileDecls = append(fileDecls, &edit.Decl{
											Kind:  edit.DeclKindVar,
											Ident: name,
											Type:  spec.Type,
											Decl:  decl,
											File:  file,
										})
									}
								}
							}
						}
					case token.CONST:
						// TODO
					case token.TYPE:
						// type Some...
						if opts.RecordTypes && !pkg.HasTypeDecls {
							for _, spec := range decl.Specs {
								spec, ok := spec.(*ast.TypeSpec)
								if !ok {
									continue
								}
								if isBlankName(spec.Name.Name) {
									continue
								}
								fileDecls = append(fileDecls, &edit.Decl{
									Kind:  edit.DeclKindType,
									Ident: spec.Name,
									Type:  spec.Type,
									Decl:  decl,
									File:  file,
								})
							}
						}
					}
				}
			}
			for _, decl := range fileDecls {
				decls[decl.Ident.Name] = decl
			}
			file.Decls = fileDecls
		}
		pkg.Decls = decls
		if opts.RecordVariables {
			pkg.HasVarDecls = true
		}
		if opts.RecordTypes {
			pkg.HasTypeDecls = true
		}
	}
}

// Traverse applies variable rewriting, and find possible
// functions `fn` such that `mock.Patch(fn)` is called.
func Traverse(packages *edit.Packages, recorder *Recorder) error {
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
