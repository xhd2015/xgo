package instrument_xgo_runtime

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/config"
	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/overlay"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/strutil"
)

//go:embed runtime_link_template_legacy_1_1_0.go
var legacyRuntimeLinkTemplate string

// Deprecated: we can remove once xgo/runtime v1.1.0 no longer used
func addLegacyFunctabInit(funcTabPkg *edit.Package, overrideContent func(absFile overlay.AbsFile, content string)) {
	// only for legacy xgo/runtime v1.1.0
	// add an extra init to also accept
	// the only risk is that, if in the future,
	// the FuncInfo changed, we need to prompt
	// the user to update the xgo/runtime
	var functabFile *edit.File
	for _, file := range funcTabPkg.Files {
		if file.File.Name == constants.FUNCTAB_FILE {
			functabFile = file
			break
		}
	}
	if functabFile != nil {
		// don't import new package
		// # github.com/xhd2015/xgo/runtime/functab
		// ../../../xgo/runtime@v1.1.0/functab/functab.go:1:69: could not import runtime (open : no such file or directory)
		// FAIL    github.com/secretflow/kuscia/pkg/datamesh/dataserver/io/builtin [build failed]
		edit := functabFile.Edit
		pos := functabFile.File.Syntax.End()
		// check if Functab Info have changed
		// insert at last to ensure maps are initialized
		patch.AddImport(edit, functabFile.File.Syntax, "__xgo_reflect", "reflect")
		lines := []string{
			"func init(){",
			"  checked:=false;",
			"  assignableTo := func(rType reflect.Type,cType reflect.Type) bool {",
			"    if rType.NumField() < cType.NumField() {",
			`      println("FuncInfo field number changed");`,
			"      return false;",
			"    };",
			"    for i := 0; i < cType.NumField(); i++ {",
			"      if rType.Field(i).Name != cType.Field(i).Name {",
			`        println("FuncInfo field ", cType.Field(i).Name," name changed");`,
			"        return false;",
			"      };",
			`      if rType.Field(i).Name == "Kind" {`,
			"        continue;",
			"      };",
			"      if !rType.Field(i).Type.AssignableTo(cType.Field(i).Type) {",
			`        println("FuncInfo field ", cType.Field(i).Name," type changed");`,
			"        return false;",
			"      };",
			"    };",
			"    return true;",
			"  };",
			"  runtime.XgoSetupRegisterHandler(func(fn unsafe.Pointer) {",
			"    if !checked {",
			"      checked = true;",
			// check if FuncInfo has changed, if changed, prompt user to upgrade xgo/runtime
			"      cType := __xgo_reflect.TypeOf(core.FuncInfo{});",
			"      rType := __xgo_reflect.TypeOf(runtime.XgoFuncInfo{});",
			"      if !assignableTo(rType, cType) {",
			`          panic("xgo: FuncInfo has changed, please upgrade:\n  go get github.com/xhd2015/xgo/runtime@latest");`,
			"      };",
			"    };",
			"    RegisterFunc((*core.FuncInfo)(fn));",
			"  });",
			"}",
		}
		edit.Insert(pos, ";"+strings.Join(lines, ""))
		overrideContent(overlay.AbsFile(functabFile.File.AbsPath), edit.Buffer().String())
	}
}

func patchLegacy(runtimeLinkDir string, overrideContent func(absFile overlay.AbsFile, content string)) error {
	trapDir := filepath.Join(filepath.Dir(runtimeLinkDir), "trap")
	err := removeLegacyVarPtrTrap(trapDir, overrideContent)
	if err != nil {
		return err
	}
	err = fixLegacyCtxRetrieval(trapDir, overrideContent)
	if err != nil {
		return err
	}
	return nil
}
func removeLegacyVarPtrTrap(trapDir string, overrideContent func(absFile overlay.AbsFile, content string)) error {
	varFile := filepath.Join(trapDir, "var.go")
	contentBytes, readErr := os.ReadFile(varFile)
	if readErr != nil {
		if config.DEBUG {
			fmt.Fprintf(os.Stderr, "failed to read legacy %s: %s\n", varFile, readErr)
		}
		return nil
	}
	content := string(contentBytes)
	idx, anchorLen, _ := strutil.SequenceOffset(content, []string{
		"func trapVarPtr(",
		"mock := stkData.getLastVarPtrMock(ptr)",
		"if mock == nil {",
		"mock = stkData.getLastVarMock(ptr)",
	}, 2, true)
	if idx < 0 {
		if config.DEBUG {
			fmt.Fprintf(os.Stderr, "cannot find legacy code to patch in %s\n", varFile)
		}
		return nil
	}
	// replace the anchor
	content = content[:idx] + "if mock == nil && false {" + content[idx+anchorLen:]
	overrideContent(overlay.AbsFile(varFile), content)
	return nil
}

const (
	legacyCtxCast = "argObj[0].valPtr.(context.Context)"
	fixedCtxCast  = "(*args[0].(*context.Context))"
)

func fixLegacyCtxRetrieval(trapDir string, overrideContent func(absFile overlay.AbsFile, content string)) error {
	buildFile := filepath.Join(trapDir, "build.go")
	contentBytes, readErr := os.ReadFile(buildFile)
	if readErr != nil {
		if config.DEBUG {
			fmt.Fprintf(os.Stderr, "failed to read legacy %s: %s\n", buildFile, readErr)
		}
		return nil
	}
	contentBytes = bytes.ReplaceAll(contentBytes, []byte(legacyCtxCast), []byte(fixedCtxCast))
	for {
		var ok bool
		contentBytes, ok = replaceContentAfter(contentBytes, []byte(recvAnchor), []byte(oldBuggyCheck), []byte(fixedCheck))
		if !ok {
			break
		}
	}
	overrideContent(overlay.AbsFile(buildFile), string(contentBytes))
	return nil
}

const (
	// 	var resObject object
	// if actRecvPtr != nil {
	recvAnchor    = "var resObject object"
	oldBuggyCheck = "if actRecvPtr != nil {"
	fixedCheck    = "if recvPtr == nil && actRecvPtr != nil {"
)

func replaceContentAfter(content []byte, anchor []byte, old []byte, new []byte) ([]byte, bool) {
	idx := bytes.Index(content, anchor)
	if idx < 0 {
		return content, false
	}
	anchorEnd := idx + len(anchor)
	oldIdx := bytes.Index(content[anchorEnd:], old)
	if oldIdx < 0 {
		return content, false
	}
	oldStart := anchorEnd + oldIdx
	newContent := make([]byte, len(content)+len(new)-len(old))
	i := 0
	copy(newContent, content[:oldStart])
	i += oldStart
	copy(newContent[i:], new)
	i += len(new)
	copy(newContent[i:], content[oldStart+len(old):])
	return newContent, true
}

func bypassVersionCheck(versionCode string) string {
	return strings.Replace(versionCode,
		"func checkVersion() error {",
		"func checkVersion() error { if true { return nil; }",
		1,
	)
}
