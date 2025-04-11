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
		hasDecls := make(map[*edit.File]struct{}, len(pkg.Files))
		for name, nameRecorder := range pkgRecorder.Names {
			if !nameRecorder.HasVarTrap {
				continue
			}
			decl := pkg.Decls[name]
			if decl == nil || decl.Kind != edit.DeclKindVar {
				continue
			}
			file := decl.File
			if file == nil {
				panic(fmt.Sprintf("decl %s.%s has no file", pkgPath, name))
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
			hasDecls[file] = struct{}{}
		}

		for file := range hasDecls {
			file.Edit.Insert(file.File.Syntax.Name.End(),
				";import "+constants.RUNTIME_PKG_NAME_VAR+` "runtime"`+
					";import "+constants.UNSAFE_PKG_NAME_VAR+` "unsafe"`,
			)
		}
	}
	return nil
}
