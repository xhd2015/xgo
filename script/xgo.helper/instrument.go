package main

import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
)

const interceptCompile = `
   var dlvDebug bool
   xgoDebugPkg := os.Getenv("XGO_HELPER_DEBUG_PKG")
	if xgoDebugPkg!="" {
		var pkg string
		for i, arg := range cmdline {
			if arg == "-p" {
				if i+1 < len(cmdline) {
					pkg = cmdline[i+1]
					break
				}
			}
		}
		// fmt.Printf("debug pkg: %s\n",os.Getenv("XGO_HELPER_DEBUG_PKG"))
		if pkg == xgoDebugPkg {
			dlvDebug = true
			// my compile
			base := filepath.Base(cmdline[0])
			if base == "compile" {
				cmdline[0] = filepath.Join(filepath.Dir(cmdline[0]), "compile.debug")
				oldCmdline := cmdline
				cmdline = []string{
					"dlv",
					"exec",
					"--listen=:2346",
					"--api-version=2",
					"--check-go-version=false",
					"--headless",
					"--",
				}
				cmdline = append(cmdline, oldCmdline...)
			}
		}
	}
`

const pipeoutput = `
	if dlvDebug {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, os.Stdout)
		cmd.Stderr = io.MultiWriter(cmd.Stderr, os.Stderr)
	}
`

func instrumentGc(goroot string) error {
	shellGo := filepath.Join(goroot, "src", "cmd", "go", "internal", "work", "shell.go")

	err := patch.EditFile(shellGo, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin intercept_compile>*/",
			"/*<end intercept_compile>*/",
			[]string{
				"\nfunc (sh *Shell) runOut(dir string",
				"cmdline :=",
				"\n",
			},
			2,
			patch.UpdatePosition_After,
			interceptCompile,
		)
		content = patch.UpdateContent(content,
			"/*<begin pipeoutput>*/",
			"/*<end pipeoutput>*/",
			[]string{
				"\nfunc (sh *Shell) runOut(dir string",
				"err = cmd.Run()",
			},
			1,
			patch.UpdatePosition_Before,
			pipeoutput,
		)

		return content, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func instrumentPkgLoad(goroot string) error {
	pkgFile := filepath.Join(goroot, "src", "cmd", "go", "internal", "load", "pkg.go")
	err := patch.EditFile(pkgFile, func(content string) (string, error) {
		content = instrument_loadPackageData_AddRuntimeImports(content)

		return content, nil
	})
	if err != nil {
		return err
	}
	return nil
}

// but this seems does not help at all
// works with go1.24
var modLoadCode = []string{
	`if os.Getenv("XGO_ALLOW_MOD_LOAD_ADD_RUNTIME_IMPORTS") == "true" {`,
	`    add := func(imports []string,pkg string) []string {`,
	`        if pkg == "runtime" || pkg == "unsafe" || pkg == "internal" || strings.HasPrefix(pkg, "runtime/") || strings.HasPrefix(pkg, "internal/") {`,
	`            return imports;`,
	`        };`,
	`       for _,imp := range imports {`,
	`           if imp == pkg {`,
	`               return imports;`,
	`           };`,
	`       };`,
	`       return append(imports, pkg);`,
	`    };`,
	`    imports = add(imports, "runtime");`,
	`    imports = add(imports, "unsafe");`,
	`    testImports = add(testImports, "runtime");`,
	`    testImports = add(testImports, "unsafe");`,
	`};`,
}

func instrumentModLoad(goroot string) error {
	// src/cmd/go/internal/modload/load.go
	modFile := filepath.Join(goroot, "src", "cmd", "go", "internal", "modload", "load.go")
	err := patch.EditFile(modFile, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin instrument_modload>*/",
			"/*<end instrument_modload>*/",
			[]string{
				"\nfunc (ld *loader) load(ctx context.Context,",
				"imports, testImports, err =",
				"if err != nil {",
				"return",
				"}",
			},
			4,
			patch.UpdatePosition_After,
			";"+strings.Join(modLoadCode, "\n"),
		)
		return content, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func instrument_loadPackageData_AddRuntimeImports(content string) string {
	// func loadPackageData(ctx context.Context,
	return patch.UpdateContent(content,
		"/*<begin add_runtime_imports>*/",
		"/*<end add_runtime_imports>*/",
		[]string{
			"\nfunc loadPackageData(ctx context.Context,",
			"return p, loaded, err",
			"\n}",
		},
		1,
		patch.UpdatePosition_Before,
		strings.Join([]string{
			`if os.Getenv("XGO_ALLOW_ADD_RUNTIME_IMPORTS") == "true" {`,
			`    if p.ImportPath != "runtime" && p.ImportPath != "internal" && p.ImportPath != "unsafe" && !strings.HasPrefix(p.ImportPath, "runtime/") && !strings.HasPrefix(p.ImportPath, "internal/") {`,
			`       var foundRuntime bool;`,
			"       for _ ,imp := range p.Imports {",
			`           if imp == "runtime" {`,
			`               foundRuntime = true;`,
			`               break;`,
			`           };`,
			`       };`,
			`       if !foundRuntime {`,
			`           p.Imports = append(p.Imports, "runtime");`,
			`       };`,
			`    };`,
			`};`,
		}, ""),
	)
}
