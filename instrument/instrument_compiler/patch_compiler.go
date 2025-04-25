package instrument_compiler

import (
	"fmt"
	"path/filepath"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var CompilerGCMain = patch.FilePath{"src", "cmd", "compile", "internal", "gc", "main.go"}

func patchCompiler(origGoroot string, goroot string, goVersion *goinfo.GoVersion, xgoSrc string, forceReset bool, syncWithLink bool) error {
	// copy compiler internal dependencies
	err := ImportCompileInternalPatch(goroot, xgoSrc, forceReset, syncWithLink)
	if err != nil {
		return err
	}
	err = MkBuiltin(origGoroot, goroot, goVersion, RuntimeExtraDef)
	if err != nil {
		return err
	}
	err = patchCompilerInternal(goroot, goVersion)
	if err != nil {
		return err
	}
	err = PatchNoder(goroot, goVersion)
	if err != nil {
		return err
	}
	err = PatchSyntaxNode(goroot, goVersion)
	if err != nil {
		return err
	}
	return nil
}

func patchCompilerInternal(goroot string, goVersion *goinfo.GoVersion) error {
	err := patchGcMain(goroot, goVersion)
	if err != nil {
		return fmt.Errorf("patching gc main:%w", err)
	}
	return nil
}

func patchGcMain(goroot string, goVersion *goinfo.GoVersion) error {
	gcMainFile := filepath.Join(goroot, filepath.Join(CompilerGCMain...))
	return patch.EditFile(gcMainFile, func(content string) (string, error) {
		content = patch.UpdateContent(
			content,
			"/*<begin import_xgo_patch>*/",
			"/*<end import_xgo_patch>*/",
			[]string{
				"package gc",
			},
			0,
			patch.UpdatePosition_After,
			`;import xgo_patch "cmd/compile/internal/xgo_rewrite_internal/patch"`,
		)

		content = patch.UpdateContent(
			content,
			"/*<begin call_xgo_patch>*/",
			"/*<end gc_import>*/",
			[]string{
				"func Main(",
				"typecheck.InitRuntime()",
				"ssagen.InitConfig()",
			},
			2,
			patch.UpdatePosition_After,
			`;xgo_patch.Patch()`,
		)
		return content, nil
	})
}
