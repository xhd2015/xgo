package syntax

import (
	"fmt"
	"io"
	"os"
	"strings"

	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types"

	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
)

func debugSyntax(files []*syntax.File) {
	dumpSyntax := os.Getenv("XGO_DEBUG_DUMP_AST")
	if dumpSyntax == "" || dumpSyntax == "false" {
		return
	}
	dumpSyntaxFile := os.Getenv("XGO_DEBUG_DUMP_AST_FILE")

	var outFile io.Writer
	if dumpSyntaxFile != "" {
		file, err := os.Create(dumpSyntaxFile)
		if err != nil {
			panic(fmt.Errorf("dump ir: %w", err))
		}
		defer file.Close()
		outFile = file
	}

	pkgName := types.LocalPkg.Name

	if pkgName == "" {
		if len(files) > 0 {
			pkgName = files[0].PkgName.Value
		}
	}

	pkgPath := xgo_ctxt.GetPkgPath()
	namePatterns := strings.Split(dumpSyntax, ",")
	for _, file := range files {
		for _, decl := range file.DeclList {
			fnDecl, ok := decl.(*syntax.FuncDecl)
			if !ok {
				continue
			}
			var fnName string
			if fnDecl.Name != nil {
				fnName = fnDecl.Name.Value
			}
			if !xgo_ctxt.MatchAnyPattern(pkgPath, pkgName, fnName, namePatterns) {
				continue
			}

			if outFile == nil {
				syntax.Fdump(os.Stderr, fnDecl)
			} else {
				syntax.Fdump(outFile, fnDecl)
			}
		}
	}
}
