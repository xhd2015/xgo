package instrument_xgo_runtime

import (
	"strings"

	"github.com/xhd2015/xgo/instrument/constants"
	"github.com/xhd2015/xgo/instrument/edit"
	"github.com/xhd2015/xgo/instrument/overlay"
	"github.com/xhd2015/xgo/instrument/patch"
)

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
