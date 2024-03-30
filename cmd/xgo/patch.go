package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/patch"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/osinfo"
	"github.com/xhd2015/xgo/support/transform"
)

// assume go 1.20
// the patch should be idempotent
// the origGoroot is used to generate runtime defs, see https://github.com/xhd2015/xgo/issues/4#issuecomment-2017880791
func patchRuntimeAndCompiler(origGoroot string, goroot string, xgoSrc string, goVersion *goinfo.GoVersion, syncWithLink bool, revisionChanged bool) error {
	if goroot == "" {
		return fmt.Errorf("requires goroot")
	}
	if isDevelopment && xgoSrc == "" {
		return fmt.Errorf("requires xgoSrc")
	}
	if !isDevelopment && !revisionChanged {
		return nil
	}

	// runtime
	err := patchRuntimeAndTesting(goroot)
	if err != nil {
		return err
	}

	// compiler
	err = patchCompiler(origGoroot, goroot, goVersion, xgoSrc, revisionChanged, syncWithLink)
	if err != nil {
		return err
	}

	return nil
}

func patchRuntimeAndTesting(goroot string) error {
	err := patchRuntimeProc(goroot)
	if err != nil {
		return err
	}
	err = patchRuntimeTesting(goroot)
	if err != nil {
		return err
	}
	return nil
}

func patchRuntimeProc(goroot string) error {
	anchors := []string{
		"func main() {",
		"doInit(", "runtime_inittask", ")", // first doInit for runtime
		"doInit(", // second init for main
		"close(main_init_done)",
		"\n",
	}
	procGo := filepath.Join(goroot, "src", "runtime", "proc.go")
	err := editFile(procGo, func(content string) (string, error) {
		content = addContentAfter(content, "/*<begin set_init_finished_mark>*/", "/*<end set_init_finished_mark>*/", anchors, patch.RuntimeProcPatch)

		// goexit1() is called for every exited goroutine
		content = addContentAfter(content,
			"/*<begin add_go_exit_callback>*/", "/*<end add_go_exit_callback>*/",
			[]string{"func goexit1() {", "\n"},
			patch.RuntimeProcGoroutineExitPatch,
		)
		return content, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func patchRuntimeTesting(goroot string) error {
	testingFile := filepath.Join(goroot, "src", "testing", "testing.go")
	return editFile(testingFile, func(content string) (string, error) {
		// func tRunner(t *T, fn func(t *T)) {
		anchor := []string{"func tRunner(t *T", "{", "\n"}
		content = addContentBefore(content,
			"/*<begin declare_testing_callback>*/", "/*<end declare_testing_callback>*/",
			anchor,
			patch.TestingCallbackDeclarations,
		)
		content = addContentAfter(content,
			"/*<begin call_testing_callback>*/", "/*<end call_testing_callback>*/",
			anchor,
			patch.TestingStart,
		)
		return content, nil
	})
}
func getInternalPatch(goroot string, subDirs ...string) string {
	dir := filepath.Join(goroot, "src", "cmd", "compile", "internal", "xgo_rewrite_internal", "patch")
	if len(subDirs) > 0 {
		dir = filepath.Join(dir, filepath.Join(subDirs...))
	}
	return dir
}
func importCompileInternalPatch(goroot string, xgoSrc string, forceReset bool, syncWithLink bool) error {
	dstDir := getInternalPatch(goroot)
	if isDevelopment {
		symLink := syncWithLink
		if osinfo.FORCE_COPY_UNSYM {
			// windows: A required privilege is not held by the client.
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

	// read from embed
	err := fs.WalkDir(patchEmbed, "patch_compiler", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "patch_compiler" {
			return os.MkdirAll(dstDir, 0755)
		}
		// TODO: test on windows if "/" works
		dstPath := filepath.Join(dstDir, strings.TrimPrefix(path, "patch_compiler/"))
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		content, err := patchEmbed.ReadFile(path)
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
	runtimeDefFiles := []string{"src", "cmd", "compile", "internal", "typecheck", "_builtin", "runtime.go"}
	if goVersion.Major == 1 && goVersion.Minor <= 19 {
		if goVersion.Minor > 16 {
			// in go1.19 and below, builtin has no _ prefix
			runtimeDefFiles = []string{"src", "cmd", "compile", "internal", "typecheck", "builtin", "runtime.go"}
		} else {
			runtimeDefFiles = []string{"src", "cmd", "compile", "internal", "gc", "builtin", "runtime.go"}
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
	return nil
}

func readXgoSrc(xgoSrc string, paths []string) ([]byte, error) {
	if isDevelopment {
		srcFile := filepath.Join(xgoSrc, "runtime", filepath.Join(paths...))
		return os.ReadFile(srcFile)
	}
	return patchEmbed.ReadFile("patch_compiler/" + strings.Join(paths, "/"))
}

func replaceBuildIgnore(content []byte) ([]byte, error) {
	const buildIgnore = "//go:build ignore"

	// buggy: content = bytes.Replace(content, []byte("//go:build ignore\n"), nil, 1)
	return replaceMarkerNewline(content, []byte(buildIgnore))
}

// addRuntimeFunctions always copy file
func addRuntimeFunctions(goroot string, goVersion *goinfo.GoVersion, xgoSrc string) (updated bool, err error) {
	if false {
		// seems unnecessary
		// TODO: needs to debug to see what will happen with auto generated files
		// we need to skip when debugging

		// add debug file
		//   rational: when debugging, dlv will jump to __xgo_autogen_register_func_helper.go
		// previousely this file does not exist, making the debugging blind
		runtimeAutoGenFile := filepath.Join(goroot, "src", "runtime", "__xgo_autogen_register_func_helper.go")
		srcAutoGen := getInternalPatch(goroot, "syntax", "helper_code.go")
		err = filecopy.CopyFile(srcAutoGen, runtimeAutoGenFile)
		if err != nil {
			return false, err
		}
	}

	dstFile := filepath.Join(goroot, "src", "runtime", "xgo_trap.go")
	content, err := readXgoSrc(xgoSrc, []string{"trap_runtime", "xgo_trap.go"})
	if err != nil {
		return false, err
	}

	content, err = replaceBuildIgnore(content)
	if err != nil {
		return false, fmt.Errorf("file %s: %w", filepath.Base(dstFile), err)
	}

	// the func.entry is a field, not a function
	if goVersion.Major == 1 && goVersion.Minor <= 17 {
		entryPatch := "fn.entry() /*>=go1.18*/"
		entryPatchBytes := []byte(entryPatch)
		idx := bytes.Index(content, entryPatchBytes)
		if idx < 0 {
			return false, fmt.Errorf("expect %q in xgo_trap.go, actually not found", entryPatch)
		}
		content = bytes.ReplaceAll(content, entryPatchBytes, []byte("fn.entry"))
	}

	// func name patch
	if goVersion.Major > 1 || goVersion.Minor > 22 {
		panic("should check the implementation of runtime.FuncForPC(pc).Name() to ensure __xgo_get_pc_name is not wrapped in print format above go1.22")
	}
	if goVersion.Major > 1 || goVersion.Minor >= 21 {
		content = append(content, []byte(patch.RuntimeGetFuncName_Go121)...)
	} else if goVersion.Major == 1 {
		if goVersion.Minor >= 17 {
			// go1.17,go1.18,go1.19
			content = append(content, []byte(patch.RuntimeGetFuncName_Go117_120)...)
		}
	}

	return true, os.WriteFile(dstFile, content, 0755)
}

func addReflectFunctions(goroot string, goVersion *goinfo.GoVersion, xgoSrc string) error {
	dstFile := filepath.Join(goroot, "src", "reflect", "xgo_reflect.go")
	content, err := readXgoSrc(xgoSrc, []string{"trap_runtime", "xgo_reflect.go"})
	if err != nil {
		return err
	}

	content, err = replaceBuildIgnore(content)
	if err != nil {
		return fmt.Errorf("file %s: %w", filepath.Base(dstFile), err)
	}

	valCode, err := transformReflectValue(filepath.Join(goroot, "src", "reflect", "value.go"))
	if err != nil {
		return fmt.Errorf("transforming reflect/value.go: %w", err)
	}
	typeCode, err := transformReflectType(filepath.Join(goroot, "src", "reflect", "type.go"))
	if err != nil {
		return fmt.Errorf("transforming reflect/type.go: %w", err)
	}

	// fmt.Printf("typCode: %s\n", typeCode)

	// concat all code
	content = bytes.Join([][]byte{content, []byte(valCode), []byte(typeCode)}, []byte("\n"))
	return os.WriteFile(dstFile, content, 0755)
}

const xgoGetAllMethodByName = "__xgo_get_all_method_by_name"

func transformReflectValue(reflectValueFile string) (string, error) {
	file, err := transform.Parse(reflectValueFile)
	if err != nil {
		return "", err
	}

	fnDecl := file.GetMethodDecl("Value", "MethodByName")
	if fnDecl == nil {
		return "", fmt.Errorf("cannot find Value.MethodByName")
	}

	code, err := replaceIdent(file, fnDecl, xgoGetAllMethodByName, func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		switch idt.Name {
		case "MethodByName":
			return idt, xgoGetAllMethodByName
		case "Method": // method by index
			return idt, "__xgo_get_all_method_index"
		}

		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing MethodByName: %w", err)
	}

	methodDecl := file.GetMethodDecl("Value", "Method") // method by index
	if methodDecl == nil {
		return "", fmt.Errorf("cannot find Value.Method")
	}
	code2, err := replaceIdent(file, methodDecl, "__xgo_get_all_method_index", func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		switch idt.Name {
		case "NumMethod": // method by index
			return idt, "__xgo_get_all_method_num"
		}
		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing Method: %w", err)
	}

	codef := strings.Join([]string{code, code2}, "\n")
	return codef, nil
}

func transformReflectType(reflectTypeFile string) (string, error) {
	file, err := transform.Parse(reflectTypeFile)
	if err != nil {
		return "", err
	}
	fnDecl := file.GetMethodDecl("rtype", "MethodByName")
	if fnDecl == nil {
		return "", fmt.Errorf("cannot find rtype.MethodByName")
	}
	m0, err := replaceIdent(file, fnDecl, xgoGetAllMethodByName, func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		if idt.Name == "ExportedMethods" {
			return idt, "Methods"
		} else if idt.Name == "Method" {
			return idt, "__xgo_get_all_method_index"
		}
		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing ExportedMethods: %w", err)
	}

	fnDecl2 := file.GetMethodDecl("rtype", "exportedMethods")
	if fnDecl2 == nil {
		return "", fmt.Errorf("cannot find rtype.exportedMethods")
	}

	m1, err := replaceIdent(file, fnDecl2, "__xgo_all_methods", func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		if idt.Name == "ExportedMethods" {
			return idt, "Methods"
		}
		return nil, ""
	})
	if err != nil {
		return "", err
	}

	methodDecl := file.GetMethodDecl("rtype", "Method")
	if methodDecl == nil {
		return "", fmt.Errorf("cannot find rtype.Method")
	}
	m2, err := replaceIdent(file, methodDecl, "__xgo_get_all_method_index", func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		if idt.Name == "exportedMethods" {
			return idt, "__xgo_all_methods"
		}
		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing Method: %w", err)
	}

	numA := file.GetMethodDecl("rtype", "NumMethod")
	if numA == nil {
		return "", fmt.Errorf("cannot find rtype.NumMethod")
	}
	m3, err := replaceIdent(file, numA, "__xgo_get_all_method_num", func(n ast.Node) (*ast.Ident, string) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		idt := sel.Sel
		if idt.Name == "exportedMethods" {
			return idt, "__xgo_all_methods"
		}
		return nil, ""
	})
	if err != nil {
		return "", fmt.Errorf("replacing Method: %w", err)
	}
	code := strings.Join([]string{m0, m1, m2, m3}, "\n")
	return code, nil
}

func replaceIdent(file *transform.File, fnDecl *ast.FuncDecl, replaceFuncName string, identReplacer func(n ast.Node) (*ast.Ident, string)) (string, error) {
	type replaceIdent struct {
		idt *ast.Ident
		rep string
	}
	var replaceIdents []replaceIdent
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		if n == nil {
			// post action
			return false
		}
		idt, replace := identReplacer(n)
		if idt != nil {
			replaceIdents = append(replaceIdents, replaceIdent{
				idt: idt,
				rep: replace,
			})
		}
		return true
	})
	if len(replaceIdents) == 0 {
		return "", errors.New("no replace found")
	}
	if replaceFuncName != "" {
		// replace the name
		replaceIdents = append(replaceIdents, replaceIdent{
			idt: fnDecl.Name,
			rep: replaceFuncName,
		})
	}
	// find assignment to x
	sort.Slice(replaceIdents, func(i, j int) bool {
		a := replaceIdents[i].idt
		b := replaceIdents[j].idt
		return file.Fset.Position(a.Pos()).Offset < file.Fset.Position(b.Pos()).Offset
	})

	// replace
	n := len(replaceIdents)
	baseOffset := file.Fset.Position(fnDecl.Pos()).Offset

	code := file.GetCode(fnDecl)
	for i := n - 1; i >= 0; i-- {
		rp := replaceIdents[i]
		offset := file.Fset.Position(rp.idt.Pos()).Offset - baseOffset

		var buf bytes.Buffer
		buf.Grow(len(code))
		buf.Write(code[:offset])
		buf.WriteString(rp.rep)
		buf.Write(code[offset+len(rp.idt.Name):])

		code = buf.Bytes()
		// NOTE: don't use slice append, content will be override
		if false {
			newCode := append(code[:offset:offset], []byte(rp.rep)...)
			newCode = append(newCode, code[offset+len(rp.idt.Name):]...)
			code = newCode
		}
	}
	return string(code), nil
}

// content = bytes.Replace(content, []byte("//go:build ignore\n"), nil, 1)
func replaceMarkerNewline(content []byte, marker []byte) ([]byte, error) {
	idx := bytes.Index(content, marker)
	if idx < 0 {
		return nil, fmt.Errorf("missing %s", string(marker))
	}
	idx += len(marker)
	if idx < len(content) && content[idx] == '\r' {
		idx++
	}
	if idx < len(content) && content[idx] == '\n' {
		idx++
	}
	return content[idx:], nil
}
func patchCompilerNoder(goroot string, goVersion *goinfo.GoVersion) error {
	files := []string{"src", "cmd", "compile", "internal", "noder", "noder.go"}
	var noderFiles string
	if goVersion.Major == 1 {
		minor := goVersion.Minor
		if minor == 16 {
			files = []string{"src", "cmd", "compile", "internal", "gc", "noder.go"}
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
	file := filepath.Join(goroot, "src", "cmd", "compile", "internal", "noder", "irgen.go")
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
	file := filepath.Join(goroot, "src", "cmd", "compile", "internal", "syntax", "xgo_nodes.go")
	return os.WriteFile(file, []byte("package syntax\n"+strings.Join(fragments, "\n")), 0755)
}

func patchGcMain(goroot string, goVersion *goinfo.GoVersion) error {
	file := filepath.Join(goroot, "src", "cmd", "compile", "internal", "gc", "main.go")
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
		// windows: also turn off optimization(gcflags=-N)
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
				// go1.19 and under does not hae PGO
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

func checkRevisionChanged(revisionFile string, currentRevision string) (bool, error) {
	savedRevision, err := readOrEmpty(revisionFile)
	if err != nil {
		return false, err
	}
	logDebug("current revision: %s, last revision: %s from file %s", currentRevision, savedRevision, revisionFile)
	if savedRevision == "" || savedRevision != currentRevision {
		return true, nil
	}
	return false, nil
}

func readOrEmpty(file string) (string, error) {
	version, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	s := string(version)
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\r")
	return s, nil
}

// NOTE: flagA never cause goroot to reset
func syncGoroot(goroot string, dstDir string, forceCopy bool) error {
	// check if src goroot has src/runtime
	srcRuntimeDir := filepath.Join(goroot, "src", "runtime")
	err := assertDir(srcRuntimeDir)
	if err != nil {
		return err
	}
	if !forceCopy {
		srcGoBin := filepath.Join(goroot, "bin", "go")
		dstGoBin := filepath.Join(dstDir, "bin", "go")

		srcFile, err := os.Stat(srcGoBin)
		if err != nil {
			return nil
		}
		if srcFile.IsDir() {
			return fmt.Errorf("bad goroot: %s", goroot)
		}

		dstFile, statErr := os.Stat(dstGoBin)
		if statErr != nil {
			if !os.IsNotExist(statErr) {
				return statErr
			}
		}

		if dstFile != nil && !dstFile.IsDir() && dstFile.Size() == srcFile.Size() {
			// already copied
			return nil
		}
	}

	// need copy, delete target dst dir first
	// TODO: use git worktree add if .git exists
	return filecopy.NewOptions().
		Concurrent(10).
		CopyReplaceDir(goroot, dstDir)
}

func buildInstrumentTool(goroot string, xgoSrc string, compilerBin string, compilerBuildIDFile string, execToolBin string, debugPkg string, logCompile bool, noSetup bool, debugWithDlv bool) (compilerChanged bool, toolExecFlag string, err error) {
	actualExecToolBin := execToolBin
	if !noSetup {
		// build the instrumented compiler
		err = buildCompiler(goroot, compilerBin)
		if err != nil {
			return false, "", err
		}
		compilerChanged, err = compareAndUpdateCompilerID(compilerBin, compilerBuildIDFile)
		if err != nil {
			return false, "", err
		}

		if isDevelopment {
			// build exec tool
			buildExecToolCmd := exec.Command("go", "build", "-o", execToolBin, "./exec_tool")
			buildExecToolCmd.Dir = filepath.Join(xgoSrc, "cmd")
			buildExecToolCmd.Stdout = os.Stdout
			buildExecToolCmd.Stderr = os.Stderr
			err = buildExecToolCmd.Run()
			if err != nil {
				return false, "", err
			}
		} else {
			actualExecToolBin, err = findBuiltExecTool()
			if err != nil {
				return false, "", err
			}
		}
	}

	execToolCmd := []string{actualExecToolBin, "--enable"}
	if logCompile {
		execToolCmd = append(execToolCmd, "--log-compile")
	}
	if debugPkg != "" {
		execToolCmd = append(execToolCmd, "--debug="+debugPkg)
	}
	if debugWithDlv {
		execToolCmd = append(execToolCmd, "--debug-with-dlv")
	}
	// always add trailing '--' to mark exec tool flags end
	execToolCmd = append(execToolCmd, "--")

	toolExecFlag = "-toolexec=" + strings.Join(execToolCmd, " ")
	return compilerChanged, toolExecFlag, nil
}

// find exec_tool, first try the same dir with xgo,
// but if that is not found, we can fallback to ~/.xgo/bin/exec_tool
// because exec_tool changes rarely, so it is safe to use
// an older version.
// we may add version to check if exec_tool is compatible
func findBuiltExecTool() (string, error) {
	dirName := filepath.Dir(os.Args[0])
	absDirName, err := filepath.Abs(dirName)
	if err != nil {
		return "", err
	}
	exeSuffix := osinfo.EXE_SUFFIX
	execToolBin := filepath.Join(absDirName, "exec_tool"+exeSuffix)
	_, statErr := os.Stat(execToolBin)
	if statErr == nil {
		return execToolBin, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("exec_tool not found in %s", dirName)
	}
	execToolBin = filepath.Join(homeDir, ".xgo", "bin", "exec_tool"+exeSuffix)
	_, statErr = os.Stat(execToolBin)
	if statErr == nil {
		return execToolBin, nil
	}
	return "", fmt.Errorf("exec_tool not found in %s and ~/.xgo/bin", dirName)
}
func buildCompiler(goroot string, output string) error {
	args := []string{"build"}
	if isDevelopment {
		args = append(args, "-gcflags=all=-N -l")
	}
	args = append(args, "-o", output, "./")
	cmd := exec.Command(filepath.Join(goroot, "bin", "go"), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env, err := patchEnvWithGoroot(os.Environ(), goroot)
	if err != nil {
		return err
	}
	cmd.Env = env
	cmd.Dir = filepath.Join(goroot, "src", "cmd", "compile")
	return cmd.Run()
}

func compareAndUpdateCompilerID(compilerFile string, compilerIDFile string) (changed bool, err error) {
	prevData, statErr := ioutil.ReadFile(compilerIDFile)
	if statErr != nil {
		if !errors.Is(statErr, os.ErrNotExist) {
			return false, statErr
		}
	}
	prevID := string(prevData)
	curID, err := getBuildID(compilerFile)
	if err != nil {
		return false, err
	}
	if prevID != "" && prevID == curID {
		return false, nil
	}
	err = ioutil.WriteFile(compilerIDFile, []byte(curID), 0755)
	if err != nil {
		return false, err
	}
	return true, nil
}

func getBuildID(file string) (string, error) {
	data, err := exec.Command("go", "tool", "buildid", file).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}
