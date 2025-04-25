package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/instrument/instrument_compiler"

	"github.com/xhd2015/xgo/cmd/xgo/asset"
	"github.com/xhd2015/xgo/cmd/xgo/patch"
	instrument_patch "github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var XgoRewriteInternal = instrument_compiler.XgoRewriteInternal
var XgoRewriteInternalPatch = instrument_compiler.XgoRewriteInternalPatch
var xgoNodes = instrument_compiler.SyntaxNodesFile

var CompilerGCMain = instrument_compiler.CompilerGCMain
var noderFile = instrument_compiler.NoderFile
var noderFile16 = instrument_compiler.NoderFile16
var irgenFile = _FilePath{"src", "cmd", "compile", "internal", "noder", "irgen.go"}

var compilerRuntimeDefFile = instrument_compiler.CompilerRuntimeDefFile
var compilerRuntimeDefFile18 = instrument_compiler.CompilerRuntimeDefFile18
var compilerRuntimeDefFile16 = instrument_compiler.CompilerRuntimeDefFile16

var compilerFiles = []_FilePath{
	xgoNodes,
	CompilerGCMain,
	noderFile,
	noderFile16,
	irgenFile,
	compilerRuntimeDefFile,
	compilerRuntimeDefFile18,
	compilerRuntimeDefFile16,

	type2ExprPatch.FilePath,
	type2AssignmentsPatch.FilePath,
	syntaxWalkPatch.FilePath,
	syntaxParserPatch.FilePath,
	noderWriterPatch.FilePath,
	noderExprPatch.FilePath,
	syntaxPrinterPatch.FilePath,
	syntaxExtra,
}

func patchCompilerLegacy(origGoroot string, goroot string, goVersion *goinfo.GoVersion, xgoSrc string, forceReset bool, syncWithLink bool) error {
	// copy compiler internal dependencies
	err := instrument_compiler.ImportCompileInternalPatch(goroot, xgoSrc, forceReset, syncWithLink)
	if err != nil {
		return err
	}
	runtimeDefUpdated, err := addRuntimeFunctions(goroot, goVersion, xgoSrc)
	if err != nil {
		return err
	}

	if runtimeDefUpdated {
		err = instrument_compiler.MkBuiltin(origGoroot, goroot, goVersion, patch.RuntimeExtraDef)
		if err != nil {
			return err
		}
	}

	// NOTE: not adding reflect to access any method
	if false {
		err = addReflectFunctions(goroot, goVersion, xgoSrc)
		if err != nil {
			return err
		}
	}

	err = patchCompilerInternalLegacy(goroot, goVersion)
	if err != nil {
		return err
	}
	return nil
}

// Deprecated
func patchCompilerInternalLegacy(goroot string, goVersion *goinfo.GoVersion) error {
	// src/cmd/compile/internal/noder/noder.go
	err := instrument_compiler.PatchNoder(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching noder: %w", err)
	}
	if goVersion.Major == goinfo.GO_MAJOR_1 && (goVersion.Minor == goinfo.GO_VERSION_18 || goVersion.Minor == goinfo.GO_VERSION_19) {
		err := poatchIRGenericGen(goroot, goVersion)
		if err != nil {
			return fmt.Errorf("patching generic trap: %w", err)
		}
	}
	err = instrument_compiler.PatchSyntaxNode(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching syntax node:%w", err)
	}
	err = patchGcMainOld(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching gc main:%w", err)
	}
	err = patchCompilerAstTypeCheck(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patch ast type check:%w", err)
	}
	return nil
}

// Deprecated
func patchGcMainOld(goroot string, goVersion *goinfo.GoVersion) error {
	file := filepath.Join(goroot, filepath.Join(CompilerGCMain...))
	go116AndUnder := goVersion.Major == 1 && goVersion.Minor <= goinfo.GO_VERSION_16
	go117 := goVersion.Major == 1 && goVersion.Minor == goinfo.GO_VERSION_17
	go118 := goVersion.Major == 1 && goVersion.Minor == goinfo.GO_VERSION_18
	go119 := goVersion.Major == 1 && goVersion.Minor == goinfo.GO_VERSION_19
	go119AndUnder := goVersion.Major == 1 && goVersion.Minor <= goinfo.GO_VERSION_19
	go120 := goVersion.Major == goinfo.GO_MAJOR_1 && goVersion.Minor == goinfo.GO_VERSION_20
	go121 := goVersion.Major == goinfo.GO_MAJOR_1 && goVersion.Minor == goinfo.GO_VERSION_21
	go122 := goVersion.Major == goinfo.GO_MAJOR_1 && goVersion.Minor == goinfo.GO_VERSION_22
	go123 := goVersion.Major == goinfo.GO_MAJOR_1 && goVersion.Minor == goinfo.GO_VERSION_23

	return instrument_patch.EditFile(file, func(content string) (string, error) {
		imports := []string{
			`xgo_patch "cmd/compile/internal/xgo_rewrite_internal/patch"`,
			`xgo_record "cmd/compile/internal/xgo_rewrite_internal/patch/record"`,
		}

		content = instrument_patch.AddCodeAfterImportsLegacy(content,
			"/*<begin gc_import>*/", "/*<end gc_import>*/",
			imports,
		)
		initRuntimeTypeCheckGo117 := `typecheck.InitRuntime()`

		var beforePatchContent string
		var patchAnchors []string

		if go116AndUnder {
			// go1.16 is pretty old
			patchAnchors = []string{
				"loadsys()",
				"parseFiles(flag.Args())",
				"finishUniverse()",
				"recordPackageName()",
			}
		} else {
			patchAnchors = []string{`noder.LoadPackage(flag.Args())`, `dwarfgen.RecordPackageName()`}
			if !go117 {
				patchAnchors = append(patchAnchors, `ssagen.InitConfig()`)
			} else {
				// go 1.17 needs to call typecheck.InitRuntime() before patch
				beforePatchContent = initRuntimeTypeCheckGo117 + "\n"
			}
		}
		patchAnchors = append(patchAnchors, "\n")
		content = instrument_patch.AddContentAfter(content,
			"/*<begin patch>*/", "/*<end patch>*/",
			patchAnchors,
			`	// insert trap points
		if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
		    `+beforePatchContent+`xgo_patch.Patch()
		}
`)

		if go117 {
			// go1.17 needs to adjust typecheck.InitRuntime before patch
			content = instrument_patch.ReplaceContentAfter(content,
				"/*<begin patch_init_runtime_type>*/", "/*<end patch_init_runtime_type>*/",
				[]string{`escape.Funcs(typecheck.Target.Decls)`, `if base.Flag.CompilingRuntime {`, "}", "\n"},
				initRuntimeTypeCheckGo117,
				`if os.Getenv("XGO_COMPILER_ENABLE")!="true" {
					`+initRuntimeTypeCheckGo117+`
				}`,
			)
		}

		// turn off inline when there is rewrite(gcflags=-l)
		// Windows: also turn off optimization(gcflags=-N)
		var flagNSwitch = ""
		if runtime.GOOS == "windows" {
			flagNSwitch = "\n" + "base.Flag.N = 1"
		}

		// there are two ways to turn off inline
		// - 1. by not calling to inline.InlinePackage
		// - 2. by override base.Flag.LowerL to 0
		// prefer 1 because it is more focused
		if go116AndUnder {
			inlineGuard := `if Debug.l != 0 {`
			inlineAnchors := []string{
				`fninit(xtop)`,
				`Curfn = nil`,
				`// Phase 5: Inlining`,
				`if Debug_typecheckinl != 0 {`,
				"\n",
			}
			content = instrument_patch.ReplaceContentAfter(content,
				"/*<begin prevent_inline>*/", "/*<end prevent_inline>*/",
				inlineAnchors,
				inlineGuard,
				`	// NOTE: turn off inline if there is any rewrite
		`+strings.TrimSuffix(inlineGuard, " {")+` && !xgo_record.HasRewritten() {`+flagNSwitch)
		} else if go117 || go118 || go119 || go120 || go121 {
			inlineCall := `inline.InlinePackage(profile)`
			if go119AndUnder {
				// go1.19 and under does not have PGO
				inlineCall = `inline.InlinePackage()`
			}
			// go1.20 does not respect rewritten content when inlined
			content = instrument_patch.ReplaceContentAfter(content,
				"/*<begin prevent_inline>*/", "/*<end prevent_inline>*/",
				[]string{`base.Timer.Start("fe", "inlining")`, `if base.Flag.LowerL != 0 {`, "\n"},
				inlineCall,
				`	// NOTE: turn off inline if there is any rewrite
		if !xgo_record.HasRewritten() {
			`+inlineCall+`
		}else{`+flagNSwitch+`
		}
`)
		} else if go122 || go123 {
			// go1.22 also does not respect rewritten content when inlined
			// NOTE: the override of LowerL is inserted after xgo_patch.Patch()
			content = instrument_patch.AddContentAfter(content,
				"/*<begin prevent_inline_by_override_flag>*/", "/*<end prevent_inline_by_override_flag>*/",
				[]string{`if base.Flag.LowerL <= 1 {`, `base.Flag.LowerL = 1 - base.Flag.LowerL`, "}", "xgo_patch.Patch()", "}", "\n"},
				`	// NOTE: turn off inline if there is any rewrite
						if xgo_record.HasRewritten() {`+flagNSwitch+`
							base.Flag.LowerL = 0
						}
				`)
		} else {
			return "", fmt.Errorf("inline for %v not defined", goVersion)
		}

		return content, nil
	})
}

func poatchIRGenericGen(goroot string, goVersion *goinfo.GoVersion) error {
	file := irgenFile.JoinPrefix(goroot)
	return instrument_patch.EditFile(file, func(content string) (string, error) {
		imports := []string{
			`xgo_patch "cmd/compile/internal/xgo_rewrite_internal/patch"`,
		}
		if goVersion.Major == 1 && goVersion.Minor >= 19 {
			imports = append(imports, `"os"`)
		}
		content = instrument_patch.AddCodeAfterImportsLegacy(content,
			"/*<begin irgen_autogen_import>*/", "/*<end irgen_autogen_import>*/",
			imports,
		)
		content = instrument_patch.AddContentAfter(content, "/*<begin irgen_generic_trap_autogen>*/", "/*<end irgen_generic_trap_autogen>*/", []string{
			`func (g *irgen) generate(noders []*noder) {`,
			`types.DeferCheckSize()`,
			`base.ExitIfErrors()`,
			`typecheck.DeclareUniverse()`,
			"\n",
		},
			patch.GenericTrapForGo118And119)
		return content, nil
	})
}

// according to https://pkg.go.dev/embed
//
//	'separator is a forward slash, even on Windows systems'
func joinEmbedPath(paths []string) string {
	return strings.Join(paths, "/")
}
func concatEmbedPath(a string, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + "/" + b
}

func readXgoSrc(xgoSrc string, paths []string) ([]byte, error) {
	if isDevelopment {
		srcFile := filepath.Join(xgoSrc, "patch", filepath.Join(paths...))
		return os.ReadFile(srcFile)
	}
	return asset.CompilerPatchGenFS.ReadFile(asset.CompilerPatchGen + "/" + strings.Join(paths, "/"))
}
