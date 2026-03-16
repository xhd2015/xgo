package instrument_unifiedtest

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

//go:embed xgo_testinfo.go
var xgoTestinfoTemplate string

//go:embed xgo_testunified.go
var xgoTestunifiedTemplate string

// src/cmd/go/internal/test
var internalTestPath = patch.FilePath{"src", "cmd", "go", "internal", "test"}

// src/cmd/go/internal/test/test.go
var testFile = patch.FilePath{"src", "cmd", "go", "internal", "test", "test.go"}

func Unify(goroot string, goVersion *goinfo.GoVersion) error {
	// src/cmd/go/internal/test
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 26) {
		return fmt.Errorf("unified test unsupported version: go%d.%d, available: go1.17~go1.26", goVersion.Major, goVersion.Minor)
	}

	err := copyFiles(internalTestPath.JoinPrefix(goroot), goVersion)
	if err != nil {
		return err
	}
	err = instrumentTestUnifyAndCleanup(goroot, goVersion)
	if err != nil {
		return err
	}

	err = instrumentLoadTest(goroot, goVersion)
	if err != nil {
		return err
	}

	err = instrumentTest2json(goroot, goVersion)
	if err != nil {
		return err
	}

	return nil
}

func copyFiles(targetDir string, goVersion *goinfo.GoVersion) error {
	err := copyXgoTestinfo(targetDir)
	if err != nil {
		return err
	}
	err = copyXgoTestunified(targetDir, goVersion)
	if err != nil {
		return err
	}
	return nil
}

func copyXgoTestinfo(targetDir string) error {
	xgoTestinfo, err := patch.RemoveBuildIgnore(xgoTestinfoTemplate)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(targetDir, "xgo_testinfo.go"), []byte(xgoTestinfo), 0644)
	if err != nil {
		return err
	}
	return nil
}

func copyXgoTestunified(targetDir string, goVersion *goinfo.GoVersion) error {
	xgoTestunified, err := patch.RemoveBuildIgnore(xgoTestunifiedTemplate)
	if err != nil {
		return err
	}
	// In go1.26, load.PackagesAndErrors gained a *modload.State param.
	// We add a package-level var to hold the state, set from the patched runTest code.
	if goVersion.Minor >= 26 {
		xgoTestunified = strings.Replace(xgoTestunified,
			`"cmd/go/internal/load"`,
			"\"cmd/go/internal/load\"\n\t\"cmd/go/internal/modload\"",
			1,
		)
		xgoTestunified = strings.Replace(xgoTestunified,
			"var _ = builderTest",
			"var _ = builderTest\nvar xgoModuleLoaderState *modload.State",
			1,
		)
		xgoTestunified = strings.ReplaceAll(xgoTestunified,
			"load.PackagesAndErrors(ctx, load.PackageOpts{}, missing)",
			"load.PackagesAndErrors(xgoModuleLoaderState, ctx, load.PackageOpts{}, missing)",
		)
	}
	err = os.WriteFile(filepath.Join(targetDir, "xgo_testunified.go"), []byte(xgoTestunified), 0644)
	if err != nil {
		return err
	}
	return nil
}

func instrumentTestUnifyAndCleanup(goroot string, goVersion *goinfo.GoVersion) error {
	fileName := testFile.JoinPrefix()
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 26) {
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.26", fileName, goVersion.Major, goVersion.Minor)
	}

	return patch.EditFile(filepath.Join(goroot, fileName), func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin defer_xgoCleanup>*/",
			"/*<end defer_xgoCleanup>*/",
			[]string{
				"func runTest(ctx context.Context,",
				") {",
			},
			1,
			patch.UpdatePosition_After,
			"defer xgoCleanup();",
		)
		// In go1.26, load.PackagesAndErrors signature changed to include moduleLoaderState
		pkgsAnchor := "pkgs = load.PackagesAndErrors(ctx,"
		if goVersion.Minor >= 26 {
			pkgsAnchor = "pkgs = load.PackagesAndErrors(moduleLoaderState, ctx,"
		}
		unifyCall := ";pkgs = xgoUnifyTestPackages(ctx, pkgs)"
		if goVersion.Minor >= 26 {
			// Set the module loader state before unifying test packages
			unifyCall = ";xgoModuleLoaderState = moduleLoaderState;pkgs = xgoUnifyTestPackages(ctx, pkgs)"
		}
		content = patch.UpdateContent(content,
			"/*<begin call_xgoUnifyTestPackages>*/",
			"/*<end call_xgoUnifyTestPackages>*/",
			[]string{
				"func runTest(ctx context.Context,",
				pkgsAnchor,
				")",
			},
			2,
			patch.UpdatePosition_After,
			unifyCall,
		)

		// set pkg dir
		runnerAnchor := "func (r *runTestActor) Act("
		if goVersion.Major == 1 && goVersion.Minor <= 19 {
			runnerAnchor = "func (c *runCache) builderRunTest("
		}
		content = patch.UpdateContent(content,
			"/*<begin set pkg dir>*/",
			"/*<end set pkg dir>*/",
			[]string{
				runnerAnchor,
				"cmd.Dir = a.Package.Dir",
			},
			1,
			patch.UpdatePosition_After,
			strings.Join([]string{
				`;if xgoCmdDir:=xgoGetCmdDir(a.Package); xgoCmdDir != "" {`,
				`cmd.Dir = xgoCmdDir;`,
				`}`,
			}, "\n"),
		)
		return content, nil
	})
}
