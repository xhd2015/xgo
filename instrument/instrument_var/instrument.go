package instrument_var

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
)

type PkgScopeNames map[string]*edit.Decl

func Instrument(packages *edit.Packages) error {
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
					fileEdit := file.Edit
					end := decl.Decl.End()
					// TODO: check if the end is a semicolon
					// `;;` causes error
					// endOffset := fset.Position(end).Offset

					infoVar := fmt.Sprintf("%s_%d_%d", constants.VAR_INFO, file.Index, len(file.TrapVars))

					declType := decl.Type
					typeStart := fset.Position(declType.Pos()).Offset
					typeEnd := fset.Position(declType.End()).Offset
					typeCode := file.File.Content[typeStart:typeEnd]

					varName := decl.Ident.Name
					code := genCode(varName, infoVar, typeCode)

					file.TrapVars = append(file.TrapVars, &edit.VarInfo{
						InfoVar: infoVar,
						Name:    varName,
						Type:    declType,
						Decl:    decl,
					})

					fileEdit.Insert(end, ";")
					fileEdit.Insert(end, code)
					hasVar = true
				}
			}
			if hasVar {
				file.Edit.Insert(file.File.Syntax.Name.End(),
					";import "+constants.RUNTIME_PKG_NAME_VAR+` "runtime"`+
						";import "+constants.UNSAFE_PKG_NAME_VAR+` "unsafe"`,
				)
			}
		}
	}
	return nil
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
