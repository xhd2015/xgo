package inject

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

// InjectRuntimeTrap parses the given file as golang AST,
// and then for each package level function decl that has a body,
// it inserts a `defer runtime.XgoTrap()();` at the beginning of the body.
// Returns the modified content.
// insert runtime.XgoTrap(), example:
//
//	func add(a, b int) int {
//		return a+b
//	}
//
//  -->
//
//	func add(a, b int) int {defer runtime.XgoTrap()();
//		return a+b
//	}

func InjectRuntimeTrap(filePath string) ([]byte, bool, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false, err
	}

	// Create the file set
	fset := token.NewFileSet()

	// Parse the file
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	// Create a new editor - convert content to string for goedit.New
	editor := goedit.New(fset, string(content))

	var hasTrap bool
	// Visit all nodes in the AST
	ast.Inspect(file, func(n ast.Node) bool {
		// Check if this is a function declaration
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			// Only process functions with a body
			if funcDecl.Body != nil {
				// Get position right after the opening brace
				pos := funcDecl.Body.Lbrace + 1

				// Insert the trap statement with a semicolon
				editor.Insert(pos, "defer __xgo_trap_runtime.XgoTrap()();")
				hasTrap = true
			}
		}
		return true
	})
	if !hasTrap {
		return nil, false, nil
	}

	editor.Insert(file.Name.End(), `;import __xgo_trap_runtime "runtime"`)

	// Return the modified content - convert string back to []byte
	return []byte(editor.String()), true, nil
}
