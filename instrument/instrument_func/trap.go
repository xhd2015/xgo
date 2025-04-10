package instrument_func

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	astutil "github.com/xhd2015/xgo/instrument/ast"
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/resolve"
	"github.com/xhd2015/xgo/support/edit/goedit"
)

const (
	recvNamePrefix   = "__xgo_auto_rcv_"
	paramNamePrefix  = "__xgo_auto_param_"
	resultNamePrefix = "__xgo_auto_res_"
)

type Options struct {
	NoFilterStdlib bool
	PkgRecorder    *resolve.PkgRecorder
}

// TrapFuncs parses the given file as golang AST,
// and then for each package level function decl that has a body,
// it inserts a `defer runtime.XgoTrap()();` at the beginning of the body.
// Returns the modified content.
// insert runtime.XgoTrap(), example:
//
//		func add(a, b int) int {
//			return a+b
//		}
//
//	 -->
//
//		func add(a, b int) int {defer runtime.XgoTrap()();
//			return a+b
//		}
func TrapFuncs(editor *goedit.Edit, pkgPath string, file *ast.File, fileIndex int, opts Options) []*edit.FuncInfo {
	fset := editor.Fset()

	recorder := opts.PkgRecorder

	var cfg stdPkgConfig
	var cfgOk bool
	if !opts.NoFilterStdlib {
		cfg, cfgOk = stdPkgConfigMapping[pkgPath]
	}

	var funcInfos []*edit.FuncInfo
	// Visit all decls in the AST
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if funcDecl.Body == nil {
			continue
		}
		if funcDecl.Name == nil || funcDecl.Name.Name == "" || funcDecl.Name.Name == "_" {
			continue
		}
		// Check if it's a method (has a receiver) with empty or "_" receiver name
		_, receiver := processReceiverNames(funcDecl, fset, editor)
		funcName := funcDecl.Name.Name
		if receiver == nil && funcName == "init" {
			continue
		}
		if pkgPath == "time" && (funcName == constants.XGO_REAL_NOW || funcName == constants.XGO_REAL_SLEEP) {
			// certain function is specifically left for xgo to call
			continue
		}
		identityName, recvPtr, recvGeneric, recvType := ParseReceiverInfo(funcName, receiver)
		if cfgOk && !cfg.whitelistFunc[identityName] && !matchAnyPrefix(cfg.whitelistFuncPrefix, identityName) {
			// TODO: may enforce only exporeted function on standard lib?
			continue
		}

		if recorder != nil {
			var hasFnRecord bool
			var hasTypeMethodRecord bool
			fnRecorder := recorder.Get(funcName)
			if fnRecorder != nil && fnRecorder.HasMockRef {
				hasFnRecord = true
			}
			if !hasFnRecord && recvType != nil {
				typeRecorder := recorder.Get(recvType.Name)
				if typeRecorder != nil && typeRecorder.NamesHavingMock[funcName] {
					hasTypeMethodRecord = true
				}
			}
			if !hasFnRecord && !hasTypeMethodRecord {
				continue
			}
		}

		_, receiverAddr := toNameAddr(receiver)
		// Process parameter names
		_, paramFields := processParamNames(funcDecl, fset, editor)

		_, resultFields := processResultNames(funcDecl, fset, editor)

		_, paramAddrs := toNameAddrs(paramFields)
		_, resultAddrs := toNameAddrs(resultFields)

		// Only process functions with a body
		// Get position right after the opening brace
		pos := funcDecl.Body.Lbrace + 1
		line := fset.Position(pos).Line

		// Insert the trap statement with a semicolon:
		//     post, stop := XgoTrap(recvName, &recv,argNames, &args,resultNames, &results)
		//     if post != nil {
		//          defer post()
		//     }
		//     if stop {
		//        return
		//     }
		// trap: func(recvName string, recvPtr interface{}, argNames []string, args []interface{}, resultNames []string, results []interface{}) (func(), bool)
		funcInfo := fmt.Sprintf("%s_%d_%d", constants.FUNC_INFO, fileIndex, len(funcInfos))
		editor.Insert(pos, fmt.Sprintf(trapTemplate,
			line, line,
			funcInfo,
			receiverAddr,
			strings.Join(paramAddrs, ","),
			strings.Join(resultAddrs, ","),
			line, line, line,
		))

		funcInfos = append(funcInfos, &edit.FuncInfo{
			InfoVar:      funcInfo,
			FuncDecl:     funcDecl,
			IdentityName: identityName,
			RecvPtr:      recvPtr,
			RecvGeneric:  recvGeneric,
			RecvType:     recvType,
			Receiver:     receiver,
			Params:       paramFields,
			Results:      resultFields,
		})
	}

	if len(funcInfos) == 0 {
		return nil
	}

	editor.Insert(file.Name.End(),
		`;import `+constants.RUNTIME_PKG_NAME_FUNC+` "runtime"`+
			`;import `+constants.UNSAFE_PKG_NAME_FUNC+` "unsafe"`,
	)
	return funcInfos
}

func ParseReceiverInfo(fnName string, receiver *edit.Field) (identityName string, recvPtr bool, recvGeneric bool, recvType *ast.Ident) {
	if receiver == nil {
		identityName = fnName
		return
	}
	return astutil.ParseReceiverInfo(fnName, receiver.Type)
}

func matchAnyPrefix(list []string, s string) bool {
	for _, p := range list {
		if p != "" && strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func toNameAddr(name *edit.Field) (string, string) {
	if name == nil || name.Name == "" {
		return `""`, "nil"
	}
	return strconv.Quote(name.Name), "&" + name.Name
}

func toNameAddrs(names []*edit.Field) ([]string, []string) {
	varNames := make([]string, len(names))
	addrs := make([]string, len(names))
	for i, name := range names {
		varNames[i] = strconv.Quote(name.Name)
		addrs[i] = "&" + name.Name
	}
	return varNames, addrs
}

// processReceiverNames processes a function declaration's receiver list,
// adding names to unnamed receivers or replacing "_" receivers with unique names.
// Returns true if any receiver names were amended.
func processReceiverNames(funcDecl *ast.FuncDecl, fset *token.FileSet, editor *goedit.Edit) (bool, *edit.Field) {
	modified, fieldNames := processFieldNames(funcDecl.Recv, recvNamePrefix, editor, fset, false, funcDecl)
	if len(fieldNames) == 0 {
		return false, nil
	}
	if len(fieldNames) > 1 {
		pos := fset.Position(funcDecl.Pos())
		panic(fmt.Sprintf("multiple receiver names at %s:%d %s", pos.Filename, pos.Line, funcDecl.Name.Name))
	}

	return modified, fieldNames[0]
}

// processParamNames processes a function declaration's parameter list using the common processFieldNames function.
func processParamNames(funcDecl *ast.FuncDecl, fset *token.FileSet, editor *goedit.Edit) (modified bool, paramNames []*edit.Field) {
	return processFieldNames(funcDecl.Type.Params, paramNamePrefix, editor, fset, false, funcDecl)
}

// processResultNames processes a function declaration's result list using the common processFieldNames function.
func processResultNames(funcDecl *ast.FuncDecl, fset *token.FileSet, editor *goedit.Edit) (modified bool, resultNames []*edit.Field) {
	return processFieldNames(funcDecl.Type.Results, resultNamePrefix, editor, fset, true, funcDecl)
}

// processFieldNames is a common function for processing parameter or result names.
// It adds names to unnamed fields or replaces "_" fields with unique names.
// Returns true if any field names were modified and the list of field names.
func processFieldNames(fieldList *ast.FieldList, namePrefix string, editor *goedit.Edit, fset *token.FileSet, isResult bool, funcDecl *ast.FuncDecl) (modified bool, fieldNames []*edit.Field) {
	// No fields
	if fieldList == nil || len(fieldList.List) == 0 {
		return false, nil
	}

	fieldNames = make([]*edit.Field, 0, len(fieldList.List))
	modified = false
	seqNum := 0

	// Check if we have a single unnamed return value for results
	singleUnnamedField := isResult && len(fieldList.List) == 1 && len(fieldList.List[0].Names) == 0

	// Process each field in the list
	for _, field := range fieldList.List {
		// For fields with explicit names (could be multiple in the same field for the same type)
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				fieldName := name.Name

				// Check if the field name is empty or "_"
				if fieldName == "" || fieldName == "_" {
					// Create a new unique name based on sequential number
					newName := fmt.Sprintf("%s%d", namePrefix, seqNum)
					// Replace the unnamed field with the new name
					editor.Replace(name.Pos(), name.End(), newName)
					fieldName = newName
					modified = true
				}

				fieldNames = append(fieldNames, &edit.Field{
					Name:      fieldName,
					NameIdent: name,
					Type:      field.Type,
				})
				seqNum++
			}
		} else {
			// Handle unnamed field (e.g., func Method(int) or func Method() int)
			// Get the type position
			typePos := field.Type.Pos()

			// Create a new unique name based on sequential number
			newName := fmt.Sprintf("%s%d", namePrefix, seqNum)

			// Special handling for single unnamed return value (results only)
			if singleUnnamedField && isResult && fieldList.Opening == token.NoPos {
				// Find the position right after the closing parenthesis of the function parameters
				openPos := funcDecl.Type.Params.Closing + 1
				// Add opening parenthesis right after function parameters
				editor.Insert(openPos, " (")
				// Add closing parenthesis right before the opening brace of function body or before the next token
				var closePos token.Pos
				if funcDecl.Body != nil {
					closePos = funcDecl.Body.Lbrace
				} else {
					// If it's a function declaration without a body, use the end of the result type
					closePos = field.Type.End()
				}
				editor.Insert(closePos, ")")
			}

			// Insert the new name before the type
			editor.Insert(typePos, newName+" ")

			fieldNames = append(fieldNames, &edit.Field{
				Name:      newName,
				NameIdent: nil,
				Type:      field.Type,
			})
			modified = true
			seqNum++
		}
	}

	return modified, fieldNames
}
