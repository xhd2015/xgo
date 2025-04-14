package instrument_go

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var srcCmdGoLoadPkgPath = patch.FilePath{"src", "cmd", "go", "internal", "load", "pkg.go"}

// run `go list -deps runtime` to check all dependencies
var runtimeImportCode = []string{
	`if p.ImportPath != "runtime" && p.ImportPath != "internal" && p.ImportPath != "unsafe" && !strings.HasPrefix(p.ImportPath, "runtime/") && !strings.HasPrefix(p.ImportPath, "internal/") {`,
	`addImport("runtime", true);`,
	`};`,
}

// instrumentPkgLoad add implicit import for runtime.
// check tests at:
// - runtime/test/build/overlay_build_cache_error_with_go
// - runtime/test/build/overlay_build_cache_ok_with_xgo
// see https://github.com/xhd2015/xgo/issues/311#issuecomment-2800001350
func instrumentPkgLoad(goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 24) {
		// src/cmd/go/internal/load/pkg.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", srcCmdGoLoadPkgPath.JoinPrefix(""), goVersion.Major, goVersion.Minor)
	}
	pkgFile := srcCmdGoLoadPkgPath.JoinPrefix(goroot)
	return patch.EditFile(pkgFile, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin add_runtime_import>*/",
			"/*<end add_runtime_import>*/",
			[]string{
				"\nfunc (p *Package) load(ctx context.Context,",
				"if !opts.IgnoreImports {",
			},
			1,
			patch.UpdatePosition_After,
			strings.Join(runtimeImportCode, ""),
		)
		return content, nil
	})
}
