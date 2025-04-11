package instrument_var

import (
	"fmt"
	"go/token"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
)

// TrapVariables inserts trap points for variables
// found by resolve.Traverse
func TrapVariables(fset *token.FileSet, packages []*edit.Package) error {
	// declare getters
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			var hasVar bool
			for _, decl := range file.Decls {
				if decl.Kind == edit.DeclKindVar && decl.HasCallRewrite {
					fileEdit := file.Edit
					end := decl.Decl.End()
					// TODO: check if the end is a semicolon
					// `;;` causes error
					// endOffset := fset.Position(end).Offset

					infoVar := fmt.Sprintf("%s_%d_%d", constants.VAR_INFO, file.Index, len(file.TrapVars))

					declType := decl.Type
					var typeCode string
					if declType != nil {
						typeStart := fset.Position(declType.Pos()).Offset
						typeEnd := fset.Position(declType.End()).Offset
						typeCode = file.File.Content[typeStart:typeEnd]
					} else if decl.ResolvedValueTypeCode != "" {
						typeCode = decl.ResolvedValueTypeCode
					}
					if typeCode == "" {
						continue
					}

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
