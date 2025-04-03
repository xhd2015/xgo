package instrument_intf

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/xgo/support/instrument/constants"
	"github.com/xhd2015/xgo/support/instrument/edit"
)

// collect interface types for registration
func CollectInterfaces(file *edit.File) []*edit.InterfaceType {
	var intfTypes []*edit.InterfaceType
	for _, decl := range file.File.Syntax.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if typeSpec.Name == nil || typeSpec.Name.Name == "" || typeSpec.Name.Name == "_" {
				continue
			}
			if typeSpec.Type == nil {
				continue
			}
			intfType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			idx := len(intfTypes)
			intfTypes = append(intfTypes, &edit.InterfaceType{
				InfoVar: fmt.Sprintf("%s_%d_%d", constants.INTF_INFO, file.Index, idx),
				Name:    typeSpec.Name.Name,
				Ident:   typeSpec.Name,
				Type:    intfType,
			})
		}
	}
	return intfTypes
}
