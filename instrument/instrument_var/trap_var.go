package instrument_var

import (
	"fmt"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/resolve"
)

// TrapVariables inserts trap points for variables
// found by resolve.Traverse
func TrapVariables(packages *edit.Packages, recorder *resolve.Recorder) error {
	fset := packages.Fset
	for pkgPath, pkgRecorder := range recorder.Pkgs {
		pkg := packages.PackageByPath[pkgPath]
		if pkg == nil {
			continue
		}
		// range over files to ensure stable order
		for _, file := range pkg.Files {
			var hasTrapVar bool
			for _, decl := range file.Decls {
				if decl.Kind != edit.DeclKindVar {
					continue
				}
				nameRecord := pkgRecorder.Names[decl.Ident.Name]
				if nameRecord == nil || !nameRecord.HasVarTrap {
					continue
				}
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
				hasTrapVar = true
			}
			if hasTrapVar {
				file.Edit.Insert(file.File.Syntax.Name.End(),
					";import "+constants.RUNTIME_PKG_NAME_VAR+` "runtime"`+
						";import "+constants.UNSAFE_PKG_NAME_VAR+` "unsafe"`,
				)
			}
		}
	}
	return nil
}
