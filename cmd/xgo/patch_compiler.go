package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo"
	"github.com/xhd2015/xgo/cmd/xgo/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/osinfo"
)

var xgoRewriteInternal = _FilePath{"src", "cmd", "compile", "internal", "xgo_rewrite_internal"}
var xgoRewriteInternalPatch = append(xgoRewriteInternal, "patch")

var xgoNodes = _FilePath{"src", "cmd", "compile", "internal", "syntax", "xgo_nodes.go"}
var gcMain = _FilePath{"src", "cmd", "compile", "internal", "gc", "main.go"}
var noderFile = _FilePath{"src", "cmd", "compile", "internal", "noder", "noder.go"}
var noderFile16 = _FilePath{"src", "cmd", "compile", "internal", "gc", "noder.go"}
var irgenFile = _FilePath{"src", "cmd", "compile", "internal", "noder", "irgen.go"}

var compilerRuntimeDefFile = _FilePath{"src", "cmd", "compile", "internal", "typecheck", "_builtin", "runtime.go"}
var compilerRuntimeDefFile18 = _FilePath{"src", "cmd", "compile", "internal", "typecheck", "builtin", "runtime.go"}
var compilerRuntimeDefFile16 = _FilePath{"src", "cmd", "compile", "internal", "gc", "builtin", "runtime.go"}

var compilerFiles = []_FilePath{
	xgoNodes,
	gcMain,
	noderFile,
	noderFile16,
	irgenFile,
	compilerRuntimeDefFile,
	compilerRuntimeDefFile18,
	compilerRuntimeDefFile16,

	type2ExprPatch.FilePath,
	type2AssignmentsPatch.FilePath,
	syntaxWalkPatch.FilePath,
	synatxParserPatch.FilePath,
	noderWriterPatch.FilePath,
	noderExprPatch.FilePath,
	syntaxPrinterPatch.FilePath,
	syntaxExtra,
}

func patchCompiler(origGoroot string, goroot string, goVersion *goinfo.GoVersion, xgoSrc string, forceReset bool, syncWithLink bool) error {
	// copy compiler internal dependencies
	err := importCompileInternalPatch(goroot, xgoSrc, forceReset, syncWithLink)
	if err != nil {
		return err
	}
	runtimeDefUpdated, err := addRuntimeFunctions(goroot, goVersion, xgoSrc)
	if err != nil {
		return err
	}

	if runtimeDefUpdated {
		err = patchRuntimeDef(origGoroot, goroot, goVersion)
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

	err = patchCompilerInternal(goroot, goVersion)
	if err != nil {
		return err
	}
	return nil
}

func patchCompilerInternal(goroot string, goVersion *goinfo.GoVersion) error {
	// src/cmd/compile/internal/noder/noder.go
	err := patchCompilerNoder(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching noder: %w", err)
	}
	if goVersion.Major == 1 && (goVersion.Minor == 18 || goVersion.Minor == 19) {
		err := poatchIRGenericGen(goroot, goVersion)
		if err != nil {
			return fmt.Errorf("patching generic trap: %w", err)
		}
	}
	err = patchSynatxNode(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching syntax node:%w", err)
	}
	err = patchGcMain(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching gc main:%w", err)
	}
	err = patchCompilerAstTypeCheck(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patch ast type check:%w", err)
	}
	return nil
}

func getInternalPatch(goroot string, subDirs ...string) string {
	dir := filepath.Join(goroot, filepath.Join(xgoRewriteInternalPatch...))
	if len(subDirs) > 0 {
		dir = filepath.Join(dir, filepath.Join(subDirs...))
	}
	return dir
}

func patchSynatxNode(goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major > 1 || goVersion.Minor >= 22 {
		return nil
	}
	var fragments []string

	if goVersion.Major == 1 {
		if goVersion.Minor < 22 {
			fragments = append(fragments, patch.NodesGen)
		}
		if goVersion.Minor <= 17 {
			fragments = append(fragments, patch.Nodes_Inspect_117)
		}
	}
	if len(fragments) == 0 {
		return nil
	}
	file := filepath.Join(goroot, filepath.Join(xgoNodes...))
	return os.WriteFile(file, []byte("package syntax\n"+strings.Join(fragments, "\n")), 0755)
}

func patchGcMain(goroot string, goVersion *goinfo.GoVersion) error {
	file := filepath.Join(goroot, filepath.Join(gcMain...))
	go116AndUnder := goVersion.Major == 1 && goVersion.Minor <= 16
	go117 := goVersion.Major == 1 && goVersion.Minor == 17
	go118 := goVersion.Major == 1 && goVersion.Minor == 18
	go119 := goVersion.Major == 1 && goVersion.Minor == 19
	go119AndUnder := goVersion.Major == 1 && goVersion.Minor <= 19
	go120 := goVersion.Major == 1 && goVersion.Minor == 20
	go121 := goVersion.Major == 1 && goVersion.Minor == 21
	go122 := goVersion.Major == 1 && goVersion.Minor == 22

	return editFile(file, func(content string) (string, error) {
		imports := []string{
			`xgo_patch "cmd/compile/internal/xgo_rewrite_internal/patch"`,
			`xgo_record "cmd/compile/internal/xgo_rewrite_internal/patch/record"`,
		}
		content = addCodeAfterImports(content,
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
		content = addContentAfter(content,
			"/*<begin patch>*/", "/*<end patch>*/",
			patchAnchors,
			`	// insert trap points
		if os.Getenv("XGO_COMPILER_ENABLE")=="true" {
		    `+beforePatchContent+`xgo_patch.Patch()
		}
`)

		if go117 {
			// go1.17 needs to adjust typecheck.InitRuntime before patch
			content = replaceContentAfter(content,
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
			content = replaceContentAfter(content,
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
			content = replaceContentAfter(content,
				"/*<begin prevent_inline>*/", "/*<end prevent_inline>*/",
				[]string{`base.Timer.Start("fe", "inlining")`, `if base.Flag.LowerL != 0 {`, "\n"},
				inlineCall,
				`	// NOTE: turn off inline if there is any rewrite
		if !xgo_record.HasRewritten() {
			`+inlineCall+`
		}else{`+flagNSwitch+`
		}
`)
		} else if go122 {
			// go1.22 also does not respect rewritten content when inlined
			// NOTE: the override of LowerL is inserted after xgo_patch.Patch()
			content = addContentAfter(content,
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

func patchCompilerNoder(goroot string, goVersion *goinfo.GoVersion) error {
	files := []string(noderFile)
	var noderFiles string
	if goVersion.Major == 1 {
		minor := goVersion.Minor
		if minor == 16 {
			files = []string(noderFile16)
			noderFiles = patch.NoderFiles_1_17
		} else if minor == 17 {
			noderFiles = patch.NoderFiles_1_17
		} else if minor == 18 {
			noderFiles = patch.NoderFiles_1_17
		} else if minor == 19 {
			noderFiles = patch.NoderFiles_1_17
		} else if minor == 20 {
			noderFiles = patch.NoderFiles_1_20
		} else if minor == 21 {
			noderFiles = patch.NoderFiles_1_21
		} else if minor == 22 {
			noderFiles = patch.NoderFiles_1_21
		}
	}
	if noderFiles == "" {
		return fmt.Errorf("unsupported: %v", goVersion)
	}
	file := filepath.Join(files...)
	return editFile(filepath.Join(goroot, file), func(content string) (string, error) {
		content = addCodeAfterImports(content,
			"/*<begin file_autogen_import>*/", "/*<end file_autogen_import>*/",
			[]string{
				`xgo_syntax "cmd/compile/internal/xgo_rewrite_internal/patch/syntax"`,
				`"io"`,
			},
		)
		var anchors []string
		if goVersion.Major == 1 && goVersion.Minor <= 16 {
			anchors = []string{
				"func parseFiles(filenames []string)",
				"for _, p := range noders {",
				"localpkg.Height = myheight",
				"\n",
			}
		} else {
			anchors = []string{
				`func LoadPackage`,
				`for _, p := range noders {`,
				`base.Timer.AddEvent(int64(lines), "lines")`,
				"\n",
			}
		}
		content = addContentAfter(content, "/*<begin file_autogen>*/", "/*<end file_autogen>*/", anchors,
			noderFiles)
		return content, nil
	})
}

func poatchIRGenericGen(goroot string, goVersion *goinfo.GoVersion) error {
	file := irgenFile.Join(goroot)
	return editFile(file, func(content string) (string, error) {
		imports := []string{
			`xgo_patch "cmd/compile/internal/xgo_rewrite_internal/patch"`,
		}
		if goVersion.Major == 1 && goVersion.Minor >= 19 {
			imports = append(imports, `"os"`)
		}
		content = addCodeAfterImports(content,
			"/*<begin irgen_autogen_import>*/", "/*<end irgen_autogen_import>*/",
			imports,
		)
		content = addContentAfter(content, "/*<begin irgen_generic_trap_autogen>*/", "/*<end irgen_generic_trap_autogen>*/", []string{
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

const patchCompilerName = "patch"

func importCompileInternalPatch(goroot string, xgoSrc string, forceReset bool, syncWithLink bool) error {
	dstDir := getInternalPatch(goroot)
	if isDevelopment {
		symLink := syncWithLink
		if osinfo.FORCE_COPY_UNSYM {
			// Windows: A required privilege is not held by the client.
			symLink = false
		}
		// copy compiler internal dependencies
		err := filecopy.CopyReplaceDir(filepath.Join(xgoSrc, "patch"), dstDir, symLink)
		if err != nil {
			return err
		}

		// remove patch/go.mod
		err = os.RemoveAll(filepath.Join(dstDir, "go.mod"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		return nil
	}

	if forceReset {
		// -a causes repatch
		err := os.RemoveAll(dstDir)
		if err != nil {
			return err
		}
	} else {
		// check if already copied
		_, statErr := os.Stat(dstDir)
		if statErr == nil {
			// skip copy if already exists
			return nil
		}
	}

	patchFS := xgo.PatchFS
	// read from embed
	err := fs.WalkDir(patchFS, patchCompilerName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == patchCompilerName {
			return os.MkdirAll(dstDir, 0755)
		}
		// TODO: test on Windows if "/" works
		dstPath := filepath.Join(dstDir, path[len(patchCompilerName)+len("/"):])
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		content, err := patchFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, content, 0755)
	})
	if err != nil {
		return err
	}

	return nil
}

func readXgoSrc(xgoSrc string, paths []string) ([]byte, error) {
	if isDevelopment {
		srcFile := filepath.Join(xgoSrc, "patch", filepath.Join(paths...))
		return os.ReadFile(srcFile)
	}
	return xgo.PatchFS.ReadFile(patchCompilerName + "/" + strings.Join(paths, "/"))
}

func patchRuntimeDef(origGoroot string, goroot string, goVersion *goinfo.GoVersion) error {
	err := prepareRuntimeDefs(goroot, goVersion)
	if err != nil {
		return err
	}

	// run mkbuiltin
	cmd := exec.Command(filepath.Join(origGoroot, "bin", "go"), "run", "mkbuiltin.go")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	var dirs []string
	if goVersion.Major > 1 || (goVersion.Major == 1 && goVersion.Minor > 16) {
		dirs = []string{goroot, "src", "cmd", "compile", "internal", "typecheck"}
	} else {
		dirs = []string{goroot, "src", "cmd", "compile", "internal", "gc"}
	}
	cmd.Dir = filepath.Join(dirs...)
	cmd.Env = os.Environ()
	cmd.Env, err = patchEnvWithGoroot(cmd.Env, origGoroot)
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
func prepareRuntimeDefs(goRoot string, goVersion *goinfo.GoVersion) error {
	runtimeDefFiles := []string(compilerRuntimeDefFile)
	if goVersion.Major == 1 && goVersion.Minor <= 19 {
		if goVersion.Minor > 16 {
			// in go1.19 and below, builtin has no _ prefix
			runtimeDefFiles = []string(compilerRuntimeDefFile18)
		} else {
			runtimeDefFiles = []string(compilerRuntimeDefFile16)
		}
	}
	runtimeDefFile := filepath.Join(runtimeDefFiles...)
	fullFile := filepath.Join(goRoot, runtimeDefFile)

	extraDef := patch.RuntimeExtraDef
	return editFile(fullFile, func(content string) (string, error) {
		content = addContentAfter(content,
			`/*<begin extra_runtime_func>*/`, `/*<end extra_runtime_func>*/`,
			[]string{`var x86HasFMA bool`, `var armHasVFPv4 bool`, `var arm64HasATOMICS bool`},
			extraDef,
		)
		return content, nil
	})
}
