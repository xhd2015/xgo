package instrument_var

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/config/config_debug"
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/instrument/resolve"
	"github.com/xhd2015/xgo/instrument/resolve/types"
)

// TrapVariables inserts trap points for variables
// found by resolve.Traverse
func TrapVariables(packages *edit.Packages, recorder *resolve.Recorder) error {
	fset := packages.Fset
	for _, pkg := range packages.Packages {
		if !pkg.AllowInstrument {
			continue
		}
		var trapAll bool
		var pkgRecorder *resolve.PkgRecorder
		if pkg.Main {
			trapAll = true
		} else {
			pkgRecorder = recorder.Pkgs[pkg.LoadPackage.GoPackage.ImportPath]
			if pkgRecorder == nil || len(pkgRecorder.Names) == 0 {
				continue
			}
		}

		pkgPath := pkg.LoadPackage.GoPackage.ImportPath
		for _, file := range pkg.Files {
			impRecorder := importRecorder{}
			for _, decl := range file.Decls {
				if decl.Kind != edit.DeclKindVar || len(decl.VarRefs) == 0 {
					continue
				}
				if !trapAll {
					nameRecord := pkgRecorder.Names[decl.Ident.Name]
					if nameRecord == nil || !nameRecord.HasVarTrap {
						continue
					}
				}
				rewriteVarDefAndRefs(fset, pkgPath, file, decl, &impRecorder)
			}
		}
	}
	return nil
}

func rewriteVarDefAndRefs(fset *token.FileSet, pkgPath string, file *edit.File, decl *edit.Decl, impRecorder *importRecorder) bool {
	if config.DEBUG {
		config_debug.OnRewriteVarDefAndRefs(pkgPath, file, decl)
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
		typeCode = string(file.File.Content[typeStart:typeEnd])
	} else if decl.ResolvedValueType != nil {
		useContext := &UseContext{
			fset:        fset,
			pkgPath:     pkgPath,
			file:        file,
			pos:         decl.Ident.Pos(),
			impRecorder: impRecorder,
		}
		typeCode = useContext.useTypeInFile(decl.ResolvedValueType)
	} else if lit, ok := decl.Value.(*ast.FuncLit); ok {
		litType := lit.Type
		typeStart := fset.Position(litType.Pos()).Offset
		typeEnd := fset.Position(litType.End()).Offset
		typeCode = string(file.File.Content[typeStart:typeEnd])
	}
	if typeCode == "" {
		return false
	}

	varName := decl.Ident.Name

	// see issue https://github.com/xhd2015/xgo/issues/313
	var varPrefix string
	if strings.HasSuffix(file.File.Name, "_test.go") && strings.HasPrefix(varName, "Test") {
		varPrefix = "T_"
	}
	code := genCode(file.Index, varPrefix, varName, infoVar, typeCode)

	file.TrapVars = append(file.TrapVars, &edit.VarInfo{
		InfoVar: infoVar,
		Name:    varName,
		Decl:    decl,
	})

	fileEdit.Insert(end, ";")
	fileEdit.Insert(end, code)

	// apply edits for all refs
	// from main module
	for _, varRef := range decl.VarRefs {
		applyRewrite(varPrefix, varRef)
	}
	return true
}

func applyRewrite(prefix string, varRef *edit.VarRef) {
	fileEdit := varRef.File.Edit
	if varRef.Addr != nil {
		// delete &
		fileEdit.Delete(varRef.Addr.Pos(), varRef.Addr.X.Pos())
	}
	if prefix != "" {
		fileEdit.Insert(varRef.NameStart, prefix)
	}
	fileEdit.Insert(varRef.NameEnd, getSuffix(varRef.NeedPtr))
}

func getSuffix(isPtr bool) string {
	if isPtr {
		return "_xgo_get_addr()"
	}
	return "_xgo_get()"
}

type UseContext struct {
	fset        *token.FileSet
	pkgPath     string
	file        *edit.File
	pos         token.Pos
	impRecorder *importRecorder

	// internal
	importEdits []*ImportEdit
}

func (c *UseContext) useTypeInFile(typ types.Type) (res string) {
	defer func() {
		if r := recover(); r != nil {
			if config.DEBUG {
				fmt.Fprintf(os.Stderr, "useTypeInFile: %v\n", r)
			}
			// cancel
			// e.g. map[pkg.Type]func()
			for _, edit := range c.importEdits {
				c.impRecorder.Remove(edit.pkgPath)
			}
			res = ""
		}
	}()
	typeCode := c.doUseTypeInFile(typ)
	// apply edit if no panic
	for _, edit := range c.importEdits {
		patch.AddImport(c.file.Edit, c.file.File.Syntax, edit.ref, edit.pkgPath)
	}
	return typeCode
}

type ImportEdit struct {
	pkgPath string
	ref     string
}

func (c *UseContext) doUseTypeInFile(typ types.Type) (res string) {
	switch typ := typ.(type) {
	case types.Basic:
		return string(typ)
	case types.NamedType:
		// check if pkg path is local
		if typ.PkgPath == c.pkgPath {
			return typ.Name
		}
		// check if name is exported
		if !token.IsExported(typ.Name) {
			panic(fmt.Sprintf("unexported type: %s", typ.Name))
		}
		pos := c.fset.Position(c.pos)
		pkgRef, ok := c.impRecorder.RecordImport(typ.PkgPath, fmt.Sprintf("__xgo_var_ref_%d_%d", pos.Line, pos.Column))
		if !ok {
			c.importEdits = append(c.importEdits, &ImportEdit{
				pkgPath: typ.PkgPath,
				ref:     pkgRef,
			})
		}
		return fmt.Sprintf("%s.%s", pkgRef, typ.Name)
	case types.PtrType:
		return "*" + c.doUseTypeInFile(typ.Elem)
	case types.Signature:
		panic("TODO signature")
	case types.Tuple:
		if len(typ) == 0 {
			return ""
		}
		if len(typ) == 1 {
			return c.doUseTypeInFile(typ[0])
		}
		list := make([]string, len(typ))
		for i, t := range typ {
			list[i] = c.doUseTypeInFile(t)
		}
		return "(" + strings.Join(list, ", ") + ")"
	case types.Map:
		return fmt.Sprintf("map[%s]%s", c.doUseTypeInFile(typ.Key), c.doUseTypeInFile(typ.Value))
	case types.Slice:
		// slice, but not array
		return fmt.Sprintf("[]%s", c.doUseTypeInFile(typ.Elem))
	case types.LazyType:
		return c.doUseTypeInFile(typ())
	case types.GenericInstanceType:
		instanceParams := make([]string, len(typ.InstanceParams))
		for i, param := range typ.InstanceParams {
			instanceParams[i] = c.doUseTypeInFile(param)
		}
		return fmt.Sprintf("%s[%s]", c.doUseTypeInFile(typ.Type), strings.Join(instanceParams, ", "))
	default:
		panic(fmt.Sprintf("unsupported type: %T", typ))
	}
}

type importRecorder struct {
	// pkg -> pkgRef, pkgRef
	// is guaranteed to have no conflicts with
	// any local name
	// naming strategy: __xgo_imp_name_<line>_<col>
	// where line and col is the first position
	// that requires this import
	pkgPathToPkgRef map[string]string
}

func (c *importRecorder) RecordImport(pkgPath string, ref string) (actualRef string, exists bool) {
	if pkgPath == "" || ref == "" {
		panic("pkgPath or ref is empty")
	}
	prev, ok := c.pkgPathToPkgRef[pkgPath]
	if ok {
		return prev, true
	}
	if c.pkgPathToPkgRef == nil {
		c.pkgPathToPkgRef = make(map[string]string, 1)
	}
	c.pkgPathToPkgRef[pkgPath] = ref
	return ref, false
}

func (c *importRecorder) Remove(pkgPath string) {
	delete(c.pkgPathToPkgRef, pkgPath)
}
