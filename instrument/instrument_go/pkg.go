package instrument_go

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var srcCmdGoLoadPkgPath = patch.FilePath{"src", "cmd", "go", "internal", "load", "pkg.go"}

var checkCanAddRuntimeDep = `p.ImportPath != "runtime" && p.ImportPath != "internal" && p.ImportPath != "unsafe" && !strings.HasPrefix(p.ImportPath, "runtime/") && !strings.HasPrefix(p.ImportPath, "internal/")`

// run `go list -deps runtime` to check all dependencies
var runtimeImportCode = []string{
	`if ` + checkCanAddRuntimeDep + ` {`,
	`addImport("runtime", true);`,
	`addImport("unsafe", true);`,
	`};`,
}

var loadPackageDataAddRuntimeImport = []string{
	`if ` + checkCanAddRuntimeDep + ` {`,
	`   var foundUnsafe bool;`,
	`   var foundRuntime bool;`,
	"   for _ ,imp := range p.Imports {",
	`       if imp == "runtime" {`,
	`           foundRuntime = true;`,
	`       }else if imp == "unsafe" {`,
	`           foundUnsafe = true;`,
	`       };`,
	`       if foundRuntime && foundUnsafe {`,
	`           break;`,
	`       };`,
	`   };`,
	`   if !foundRuntime {`,
	`       p.Imports = append(p.Imports, "runtime");`,
	`   };`,
	`   if !foundUnsafe {`,
	`       p.Imports = append(p.Imports, "unsafe");`,
	`   };`,
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
		if goVersion.Minor < 19 {
			// this issue only happens in go1.19 and later
			// return cleaned content
			return content, nil
		}
		if false {
			// if we have modified loadPackageData,
			// we don't need to add runtime import here
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
		}

		returnAnchor := "return p, loaded, err"
		code := strings.Join(loadPackageDataAddRuntimeImport, "")
		if goVersion.Major == 1 && goVersion.Minor <= 20 {
			returnAnchor = "return data.p, loaded, data.err"
			code = "p:=data.p;" + code
		}
		content = patch.UpdateContent(content,
			"/*<begin add_runtime_imports_for_loadPackageData>*/",
			"/*<end add_runtime_imports_for_loadPackageData>*/",
			[]string{
				"\nfunc loadPackageData(ctx context.Context,",
				returnAnchor,
				"\n}",
			},
			1,
			patch.UpdatePosition_Before,
			code,
		)
		return content, nil
	})
}
