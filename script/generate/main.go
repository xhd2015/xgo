package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/script/build-release/revision"
	"github.com/xhd2015/xgo/script/generate/gen_defs"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/git"
	"github.com/xhd2015/xgo/support/goparse"
	"github.com/xhd2015/xgo/support/transform"
)

type GenernateType = gen_defs.GenernateType

const (
	GenernateType_DotAll              GenernateType = "./..."
	GenernateType_CompilerPatch       GenernateType = "compiler-patch"
	GenernateType_CompilerHelperCode  GenernateType = "compiler-helper-code"
	GenernateType_CompilerPatternCode GenernateType = "compiler-pattern-code"
	GenernateType_RuntimeDef          GenernateType = "runtime-def"
)

var allGenerateTypes = []GenernateType{
	gen_defs.GenernateType_CmdXgoVersion,
	gen_defs.GenernateType_RuntimeCoreVersion,
	gen_defs.GenernateType_RuntimeTraceModel,
	gen_defs.GenernateType_XgoRuntimeGen,
	gen_defs.GenernateType_XgoCompilerPatchGen,
	gen_defs.GenernateType_ScriptInstallUpgrade,
	gen_defs.GenernateType_RuntimeCoreFunc,
	gen_defs.GenernateType_RuntimeXgoTrapTemplate,
	gen_defs.GenernateType_CompilerInstrument,
	gen_defs.GenernateType_LegacyRuntimeLink,
}

func main() {
	args := os.Args[1:]

	var rootDir string
	var subGens []GenernateType
	var list bool
	var amend bool
	var noUpdateVersion bool
	if len(args) > 0 && args[0] == "list" {
		list = true
		args = args[1:]
	}
	n := len(args)
	var remainArgs []string
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if arg == "--root-dir" {
			rootDir = args[i+1]
			i++
			continue
		}
		if arg == "--list" {
			list = true
			continue
		}
		if arg == "--amend" {
			amend = true
			continue
		}
		if arg == "--no-update-version" {
			noUpdateVersion = true
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			subGens = append(subGens, GenernateType(arg))
			continue
		}
		fmt.Fprintf(os.Stderr, "unrecognized flag: %s\n", arg)
		os.Exit(1)
	}

	if list {
		for _, genType := range allGenerateTypes {
			fmt.Println(genType)
		}
		return
	}
	if len(remainArgs) > 0 {
		fmt.Fprintf(os.Stderr, "unrecognized remaining flags: %s\n", strings.Join(remainArgs, " "))
		os.Exit(1)
	}
	err := generate(rootDir, subGens, amend, noUpdateVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type SubGens []GenernateType

func (c SubGens) Has(genType GenernateType) bool {
	if len(c) == 0 {
		return true
	}
	for _, subGen := range c {
		if subGen == genType {
			return true
		}
	}
	return false
}

func generate(rootDir string, subGens SubGens, amend bool, noUpdateVersion bool) error {
	if rootDir == "" {
		resolvedRoot, err := git.ShowTopLevel("")
		if err != nil {
			return err
		}
		rootDir = resolvedRoot
	}

	if subGens.Has(gen_defs.GenernateType_CmdXgoVersion) {
		err := revision.IncrementXgoVersion(rootDir, amend, !noUpdateVersion)
		if err != nil {
			return err
		}
	}
	if subGens.Has(gen_defs.GenernateType_RuntimeCoreVersion) {
		err := revision.CopyCoreVersion(revision.GetXgoVersionFile(rootDir), revision.GetRuntimeVersionFile(rootDir))
		if err != nil {
			return err
		}
	}
	needCopyTrace := true
	if subGens.Has(gen_defs.GenernateType_RuntimeTraceModel) {
		err := copyTraceModel(rootDir)
		if err != nil {
			return err
		}
		needCopyTrace = false
	}

	if subGens.Has(gen_defs.GenernateType_RuntimeCoreFunc) {
		err := replaceCoreFunc()
		if err != nil {
			return err
		}

	}
	if subGens.Has(gen_defs.GenernateType_RuntimeXgoTrapTemplate) {
		err := replaceFuncInXgoTrap()
		if err != nil {
			return err
		}
	}

	if subGens.Has(gen_defs.GenernateType_CompilerInstrument) {
		err := copyCompilerInstrument(rootDir, string(gen_defs.GenernateType_CompilerInstrument))
		if err != nil {
			return err
		}
	}

	if subGens.Has(gen_defs.GenernateType_LegacyRuntimeLink) {
		err := replaceFuncInLegacyRuntimeLink()
		if err != nil {
			return err
		}
	}

	// gen
	if subGens.Has(gen_defs.GenernateType_XgoRuntimeGen) {
		err := genXgoRuntime(rootDir, needCopyTrace, string(gen_defs.GenernateType_XgoRuntimeGen))
		if err != nil {
			return err
		}
	}
	if subGens.Has(gen_defs.GenernateType_XgoCompilerPatchGen) {
		err := genXgoCompilerPatch(rootDir)
		if err != nil {
			return err
		}
	}
	if subGens.Has(gen_defs.GenernateType_ScriptInstallUpgrade) {
		upgradeDst := filepath.Join(rootDir, "script", "install", "upgrade")
		err := os.RemoveAll(upgradeDst)
		if err != nil {
			return err
		}

		err = copyUpgrade(filepath.Join(rootDir, "cmd", "xgo", "upgrade"), upgradeDst)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deprecated: xgo v1.1.0 has deprecated other generates
// since except the runtime_gen, other generates target
// seems vague.
func generateV1_0(rootDir string, subGens SubGens) error {
	if rootDir == "" {
		resolvedRoot, err := git.ShowTopLevel("")
		if err != nil {
			return err
		}
		rootDir = resolvedRoot
	}
	if subGens.Has(GenernateType_DotAll) {
		err := cmd.Dir(rootDir).Run("go", "generate", "./...")
		if err != nil {
			return err
		}
	}
	if subGens.Has(GenernateType_CompilerPatch) {
		err := generateCompilerPatch(rootDir)
		if err != nil {
			return err
		}
	}
	if subGens.Has(GenernateType_RuntimeDef) {
		err := generateRunTimeDefs(
			filepath.Join(rootDir, "patch", "trap_runtime", "xgo_trap.go"),
			filepath.Join(rootDir, "cmd", "xgo", "patch", "runtime_def_gen.go"),
			filepath.Join(rootDir, "patch", "syntax", "syntax_gen.go"),
			filepath.Join(rootDir, "patch", "trap_gen.go"),
		)
		if err != nil {
			return err
		}
	}
	if subGens.Has(GenernateType_CompilerPatternCode) {
		// copy files
		srcDir := filepath.Join(rootDir, "support", "pattern")
		dstDir := filepath.Join(rootDir, "patch", "match")
		files, err := os.ReadDir(srcDir)
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") || strings.HasSuffix(file.Name(), "_test.go") {
				continue
			}
			content, err := ioutil.ReadFile(filepath.Join(srcDir, file.Name()))
			if err != nil {
				return err
			}
			newContent := strings.Replace(string(content), "package pattern", "package match", 1)
			newContent = prelude + newContent
			err = ioutil.WriteFile(filepath.Join(dstDir, file.Name()), []byte(newContent), 0755)
			if err != nil {
				return err
			}
		}
	}
	if subGens.Has(GenernateType_CompilerHelperCode) {
		info, err := generateFuncHelperCode(filepath.Join(rootDir, "patch", "syntax", "helper_code.go"))
		if err != nil {
			return err
		}
		infoCode := info.formatCode("syntax")
		err = os.WriteFile(filepath.Join(rootDir, "patch", "syntax", "helper_code_gen.go"), []byte(infoCode), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

const prelude = "// Code generated by script/generate; DO NOT EDIT.\n" + "\n"

// including new line
func getCmdPrelude(cmd string) string {
	return "// Code generated by script/generate " + cmd + "; DO NOT EDIT.\n" + "\n"
}

func generateRunTimeDefs(file string, defFile string, syntaxFile string, trapFile string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	code := string(content)
	astFile, fset, err := parseGoFile(file, true)
	if err != nil {
		return err
	}
	var decls []string
	var sigRegisterFunc string
	var sigTrap string
	for _, decl := range astFile.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Recv != nil {
			continue
		}
		fnName := fn.Name.Name
		if !strings.HasPrefix(fnName, "__xgo") {
			continue
		}
		funcType := fn.Type
		if fn.Name.Name == "__xgo_register_func" {
			sigRegisterFunc = getSignature(code, fset, funcType)
		} else if fn.Name.Name == "__xgo_trap" {
			sigTrap = getSignature(code, fset, funcType)
		}
		decls = append(decls, getSlice(code, fset, funcType.Pos(), funcType.End()))
	}
	if sigRegisterFunc == "" {
		return fmt.Errorf("__xgo_register_func not found")
	}
	if sigTrap == "" {
		return fmt.Errorf("__xgo_trap not found")
	}

	declCode := "// xgo\n" + strings.Join(decls, "\n")

	defGenCode := prelude + "package patch\n" + "\n" + "//go:generate go run ../../../script/generate " + string(GenernateType_RuntimeDef) + "\n" + "const RuntimeExtraDef = `\n" + declCode + "`"
	syntaxGenCode := prelude + "package syntax\n" + "\n" + "const sig_gen__xgo_register_func = `" + sigRegisterFunc + "`"
	trapGenCode := prelude + "package patch\n" + "\n" + "const sig_gen__xgo_trap = `" + sigTrap + "`"
	err = ioutil.WriteFile(defFile, []byte(defGenCode), 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(syntaxFile, []byte(syntaxGenCode), 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(trapFile, []byte(trapGenCode), 0755)
	if err != nil {
		return err
	}

	return nil
}

type genInfo struct {
	funcStub   string
	helperCode string
}

func (c *genInfo) formatCode(pkgName string) string {
	codes := []string{
		prelude,
		fmt.Sprintf("package %s", pkgName),
	}
	codes = append(codes, fmt.Sprintf("const __xgo_stub_def = `%s`", c.funcStub))
	codes = append(codes, "")
	codes = append(codes, fmt.Sprintf("const helperCodeGen = `%s`", c.helperCode))
	codes = append(codes, "")
	return strings.Join(codes, "\n")
}

func generateFuncHelperCode(srcFile string) (*genInfo, error) {
	astFile, fset, err := parseGoFile(srcFile, true)
	if err != nil {
		return nil, err
	}
	funcStub := transform.GetTypeDecl(astFile.Decls, "__xgo_local_func_stub")
	if funcStub == nil {
		return nil, fmt.Errorf("type __xgo_local_func_stub not found")
	}

	st, ok := funcStub.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("expect __xgo_local_func_stub to be StructType, actual: %T", funcStub.Type)
	}
	codeBytes, err := os.ReadFile(srcFile)
	if err != nil {
		return nil, err
	}
	code := string(codeBytes)

	funcStubCode := getSlice(code, fset, st.Pos(), st.End())

	// TODO: use astFile.FileEnd
	var fileEnd token.Pos
	fileEnd = astFile.End()

	helperCode := getSlice(code, fset, astFile.Name.End(), fileEnd)
	return &genInfo{
		funcStub:   funcStubCode,
		helperCode: helperCode,
	}, nil
}

func copyTraceExport(srcFile string, targetFile string) error {
	contentBytes, err := os.ReadFile(srcFile)
	if err != nil {
		return err
	}
	content := string(contentBytes)
	// replace package
	content = strings.ReplaceAll(content, "package trace", "package trace")
	content = prelude + content

	return os.WriteFile(targetFile, []byte(content), 0755)
}

func copyUpgrade(srcDir string, targetDir string) error {
	err := filecopy.CopyReplaceDir(srcDir, targetDir, false)
	if err != nil {
		return err
	}
	files, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		fullFile := filepath.Join(targetDir, file.Name())
		content, err := os.ReadFile(fullFile)
		if err != nil {
			return err
		}
		content = append([]byte(prelude), content...)
		err = os.WriteFile(fullFile, content, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
func getSlice(code string, fset *token.FileSet, start token.Pos, end token.Pos) string {
	i := fset.Position(start).Offset
	j := fset.Position(end).Offset
	return code[i:j]
}
func getSignature(code string, fset *token.FileSet, funcType *ast.FuncType) string {
	var end token.Pos

	if funcType.Results != nil {
		end = funcType.Results.End()
	} else {

		end = funcType.Params.End()
	}

	return "func" + getSlice(code, fset, funcType.Params.Pos(), end)
}

func parseGoFile(file string, hasPkg bool) (*ast.File, *token.FileSet, error) {
	fileName := file
	var contentReader io.Reader
	if file == "-" {
		fileName = "<stdin>"
		contentReader = os.Stdin
	} else {
		readFile, err := os.Open(file)
		if err != nil {
			return nil, nil, err
		}
		contentReader = readFile
	}
	content, err := ioutil.ReadAll(contentReader)
	if err != nil {
		return nil, nil, err
	}
	if !hasPkg {
		contentStr := string(content)
		contentStr = goparse.AddMissingPackage(contentStr, "main")
		content = []byte(contentStr)
	}

	return goparse.ParseFileCode(fileName, content)
}
