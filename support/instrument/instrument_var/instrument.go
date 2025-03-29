package instrument_var

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/support/instrument/edit"
)

type PkgScopeNames map[string]*edit.Decl

func Instrument(packages *edit.Packages) {
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
			instrumentFunction(packages, pkg, file)
		}
	}

	fset := packages.Fset

	// declare getters
	for _, pkg := range packages.Packages {
		for _, file := range pkg.Files {
			var hasVar bool
			for _, decl := range file.Decls {
				if decl.Kind == edit.DeclKindVar && decl.Type != nil && decl.HasCallRewrite {
					edit := file.Edit
					end := decl.Decl.End()
					// TODO: check if the end is a semicolon
					// `;;` causes error
					// endOffset := fset.Position(end).Offset

					typeStart := fset.Position(decl.Type.Pos()).Offset
					typeEnd := fset.Position(decl.Type.End()).Offset
					typeCode := file.File.Content[typeStart:typeEnd]
					code := genCode(decl.Ident.Name, typeCode)

					edit.Insert(end, ";")
					edit.Insert(end, code)
					hasVar = true
				}
			}
			if hasVar {
				file.Edit.Insert(file.File.Syntax.Name.End(), `;import __xgo_var_runtime "runtime"`)
			}
		}
	}
}

func instrumentFunction(packages *edit.Packages, pkg *edit.Package, file *edit.File) {
	rootCtx := &BlockContext{
		Global: &GlobalContext{
			Packages: packages,
			Package:  pkg,
			File:     file,
		},
		Names: make(map[string]bool),
	}
	imports := getFileImports(file.File.Syntax)

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
		rootCtx.traverseFunc(fnDecl.Recv, fnDecl.Type, fnDecl.Body, pkg.Decls, imports)
	}
}
