package instrument_func

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/edit/goedit"
	"github.com/xhd2015/xgo/support/instrument/overlay"
	"github.com/xhd2015/xgo/support/instrument/patch"
)

const (
	recvNamePrefix   = "__xgo_auto_rcv_"
	paramNamePrefix  = "__xgo_auto_param_"
	resultNamePrefix = "__xgo_auto_res_"
)

// InjectRuntimeTrap parses the given file as golang AST,
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
func InjectRuntimeTrap(filePath overlay.AbsFile, overlayFS overlay.Overlay) ([]byte, bool, error) {
	_, content, err := overlayFS.Read(filePath)
	if err != nil {
		return nil, false, err
	}

	// Create the file set
	fset := token.NewFileSet()

	// Parse the file
	file, err := parser.ParseFile(fset, string(filePath), content, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	// Create a new editor - convert content to string for goedit.New
	editor := goedit.New(fset, string(content))

	hasEdited := EditInjectRuntimeTrap(editor, file)
	if !hasEdited {
		return nil, false, nil
	}
	// prefix the modified content with line directive
	editor.Buffer().Insert(0, patch.FmtLineDirective(string(filePath), 1)+"\n")
	return editor.Buffer().Bytes(), true, nil
}

func EditInjectRuntimeTrap(editor *goedit.Edit, file *ast.File) bool {
	fset := editor.Fset()
	var hasInsertedTrap bool
	var hasAmendedReceiverName bool
	var hasResNamesModified bool
	var hasParamNamesModified bool
	// Visit all nodes in the AST
	ast.Inspect(file, func(n ast.Node) bool {
		// Check if this is a function declaration
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if funcDecl.Body == nil {
			return true
		}

		// Check if it's a method (has a receiver) with empty or "_" receiver name
		var receiverName string
		if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
			modified, name := processReceiverNames(funcDecl, fset, editor)
			if modified {
				hasAmendedReceiverName = true
			}
			receiverName = name
		}

		receiverName, receiverAddr := toNameAddr(receiverName)
		// Process parameter names
		modifiedParamNames, paraNames := processParamNames(funcDecl, fset, editor)
		if modifiedParamNames {
			hasParamNamesModified = true
		}

		modifiedNames, resNames := processResultNames(funcDecl, fset, editor)
		if modifiedNames {
			hasResNamesModified = true
		}

		paramNames, paramAddrs := toNameAddrs(paraNames)
		resultNames, resultAddrs := toNameAddrs(resNames)

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
		editor.Insert(pos, fmt.Sprintf(trapTemplate,
			line, line,
			receiverName, receiverAddr,
			strings.Join(paramNames, ","), strings.Join(paramAddrs, ","),
			strings.Join(resultNames, ","), strings.Join(resultAddrs, ","),
			line, line, line,
		))
		hasInsertedTrap = true
		return true
	})
	if !hasInsertedTrap && !hasAmendedReceiverName && !hasResNamesModified && !hasParamNamesModified {
		return false
	}

	editor.Insert(file.Name.End(), `;import __xgo_trap_runtime "runtime"`)
	return true
}

func toNameAddr(name string) (string, string) {
	if name == "" {
		return `""`, "nil"
	}
	return strconv.Quote(name), "&" + name
}

func toNameAddrs(names []string) ([]string, []string) {
	varNames := make([]string, len(names))
	addrs := make([]string, len(names))
	for i, name := range names {
		varNames[i] = strconv.Quote(name)
		addrs[i] = "&" + name
	}
	return varNames, addrs
}

// processReceiverNames processes a function declaration's receiver list,
// adding names to unnamed receivers or replacing "_" receivers with unique names.
// Returns true if any receiver names were amended.
func processReceiverNames(funcDecl *ast.FuncDecl, fset *token.FileSet, editor *goedit.Edit) (bool, string) {
	var hasAmendedReceiverName bool
	var recevName string

	// Process each receiver (usually just one)
	for _, field := range funcDecl.Recv.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				// Check if the receiver name is empty or "_"
				if name.Name == "" || name.Name == "_" {
					// Get the line number for the receiver
					line := fset.Position(name.Pos()).Line
					// Create a new unique name based on the line number
					newName := fmt.Sprintf("%s%d", recvNamePrefix, line)
					// Replace the unnamed receiver with the new name
					editor.Replace(name.Pos(), name.End(), newName)
					hasAmendedReceiverName = true
					recevName = newName
				} else {
					recevName = name.Name
				}
			}
		} else {
			// Handle case where receiver has no name (e.g., func (*Type) Method())
			// Get the type position
			typePos := field.Type.Pos()
			line := fset.Position(typePos).Line
			// Create a new unique name based on the line number
			newName := fmt.Sprintf("%s%d", recvNamePrefix, line)
			// Insert the new name before the type
			editor.Insert(typePos, newName+" ")
			hasAmendedReceiverName = true
			recevName = newName
		}
	}

	return hasAmendedReceiverName, recevName
}

// processParamNames processes a function declaration's parameter list using the common processFieldNames function.
func processParamNames(funcDecl *ast.FuncDecl, fset *token.FileSet, editor *goedit.Edit) (modified bool, paramNames []string) {
	return processFieldNames(funcDecl.Type.Params, paramNamePrefix, editor, fset, false, funcDecl)
}

// processResultNames processes a function declaration's result list using the common processFieldNames function.
func processResultNames(funcDecl *ast.FuncDecl, fset *token.FileSet, editor *goedit.Edit) (modified bool, resultNames []string) {
	return processFieldNames(funcDecl.Type.Results, resultNamePrefix, editor, fset, true, funcDecl)
}

// processFieldNames is a common function for processing parameter or result names.
// It adds names to unnamed fields or replaces "_" fields with unique names.
// Returns true if any field names were modified and the list of field names.
func processFieldNames(
	fieldList *ast.FieldList,
	namePrefix string,
	editor *goedit.Edit,
	fset *token.FileSet,
	isResult bool,
	funcDecl *ast.FuncDecl,
) (modified bool, fieldNames []string) {
	// No fields
	if fieldList == nil || len(fieldList.List) == 0 {
		return false, nil
	}

	fieldNames = make([]string, 0, len(fieldList.List))
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

				fieldNames = append(fieldNames, fieldName)
				seqNum++
			}
		} else {
			// Handle unnamed field (e.g., func Method(int) or func Method() int)
			// Get the type position
			typePos := field.Type.Pos()

			// Create a new unique name based on sequential number
			newName := fmt.Sprintf("%s%d", namePrefix, seqNum)

			// Special handling for single unnamed return value (results only)
			if singleUnnamedField && isResult {
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

			fieldNames = append(fieldNames, newName)
			modified = true
			seqNum++
		}
	}

	return modified, fieldNames
}
