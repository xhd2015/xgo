package instrument_go

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

// instrument the `go` command to fix coverage with -overlay
// see https://github.com/xhd2015/xgo/issues/300

//go:embed xgo_main_template.go
var xgoMainTemplate string

var srcCmdGoPath = patch.FilePath{"src", "cmd", "go"}
var mainFilePath = srcCmdGoPath.Append("main.go")
var execFilePath = srcCmdGoPath.Append("internal", "work", "exec.go")
var xgoMainFilePath = srcCmdGoPath.Append("xgo_main.go")

func BuildInstrument(goroot string, goVersion *goinfo.GoVersion) error {
	err := instrumentExec(goroot, goVersion)
	if err != nil {
		return err
	}
	err = instrumentGoMain(goroot, goVersion)
	if err != nil {
		return err
	}
	err = copyXgoMain(goroot)
	if err != nil {
		return err
	}
	err = instrumentPkgLoad(goroot, goVersion)
	if err != nil {
		return err
	}
	// build go command
	return build.BuildBinary(goroot, filepath.Join(goroot, "src"), filepath.Join(goroot, "bin"), "go", "./cmd/go")
}

func instrumentGoMain(goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 24) {
		// src/cmd/go/main.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", execFilePath.JoinPrefix(""), goVersion.Major, goVersion.Minor)
	}
	mainFile := mainFilePath.JoinPrefix(goroot)
	return patch.EditFile(mainFile, func(content string) (string, error) {
		content = patch.UpdateContent(content,
			"/*<begin call_xgo_precheck>*/",
			"/*<end call_xgo_precheck>*/",
			[]string{
				"\nfunc main() {",
				// before the first command
				"if args[0] ==",
			},
			1,
			patch.UpdatePosition_Before,
			"if xgoPrecheck(args[0], args[1:]) { return; };",
		)
		return content, nil
	})
}

func copyXgoMain(goroot string) error {
	code, err := patch.RemoveBuildIgnore(xgoMainTemplate)
	if err != nil {
		return err
	}
	xgoMainFile := xgoMainFilePath.JoinPrefix(goroot)
	return os.WriteFile(xgoMainFile, []byte(code), 0644)
}

// instrumentExec instrument the internal exec.go to fix coverage with -overlay
func instrumentExec(goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 24) {
		// src/cmd/go/internal/work/exec.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", execFilePath.JoinPrefix(""), goVersion.Major, goVersion.Minor)
	}
	execFile := execFilePath.JoinPrefix(goroot)

	return patch.EditFile(execFile, func(content string) (string, error) {
		// since 22
		coverLine := `if p.Internal.Cover.Mode != "" {`
		getActualFile := `__xgo_overlay_source_file := sourceFile; if _actual := fsys.Actual(sourceFile); _actual != "" { __xgo_overlay_source_file = _actual; }; ` // since 24
		if goVersion.Minor < 24 {
			getActualFile = "__xgo_overlay_source_file, _ := fsys.OverlayPath(sourceFile);"
		}
		switch goVersion.Minor {
		case 20, 21:
			coverLine = `if p.Internal.CoverMode != "" {`
		case 18, 19, 17:
			coverLine = `if a.Package.Internal.CoverMode != "" {`
		}
		content = patch.UpdateContent(content,
			"/*<begin fix_cover_overlay_var_declare>*/",
			"/*<end fix_cover_overlay_var_declare>*/",
			[]string{
				"func (b *Builder) build(",
				coverLine,
				"if err := b.cover(",
			},
			2,
			patch.UpdatePosition_Before,
			getActualFile,
		)
		content = patch.UpdateContent(content,
			"/*<begin fix_cover_overlay_var_replace>*/",
			"/*<end fix_cover_overlay_var_replace>*/",
			[]string{
				"func (b *Builder) build(",
				coverLine,
				"if err := b.cover(",
				"sourceFile,",
			},
			3,
			patch.UpdatePosition_Replace,
			"__xgo_overlay_source_file,",
		)
		// redesign
		if goVersion.Minor >= 20 {
			// fmt.Fprintf(os.Stderr, "DEBUG content: \n%s\n", content)
			content = patch.UpdateContent(content,
				"/*<begin modify_infiles>*/",
				"/*<end modify_infiles>*/",
				[]string{
					"func (b *Builder) build(",
					coverLine,
					"if cfg.Experiment.CoverageRedesign {",
					"infiles = append(infiles, sourceFile)",
				},
				3,
				patch.UpdatePosition_Replace,
				strings.TrimSuffix(getActualFile, ";")+";"+"infiles = append(infiles, __xgo_overlay_source_file)",
			)
		}
		return content, nil
	})
}
