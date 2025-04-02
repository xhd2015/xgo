package instrument_go

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/goinfo"
	"github.com/xhd2015/xgo/support/instrument/patch"
	"github.com/xhd2015/xgo/support/osinfo"
)

// instrument the `go` command to fix coverage with -overlay
// see https://github.com/xhd2015/xgo/issues/300

var execFilePath = patch.FilePath{"src", "cmd", "go", "internal", "work", "exec.go"}

func InstrumentGo(goroot string, goVersion *goinfo.GoVersion) error {
	err := instrumentExec(goroot, goVersion)
	if err != nil {
		return err
	}
	// build go command
	return buildBinary(goroot, filepath.Join(goroot, "src"), filepath.Join(goroot, "bin"), "go", "./cmd/go")
}

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

func buildBinary(goroot string, srcDir string, outputDir string, outputName string, arg string) error {
	origGo := filepath.Join(goroot, "bin", "go"+osinfo.EXE_SUFFIX)

	origFile := filepath.Join(outputDir, outputName+osinfo.EXE_SUFFIX)
	tmpBuiltOutput := filepath.Join(outputDir, "__xgo_"+outputName+osinfo.EXE_SUFFIX)
	err := cmd.Dir(srcDir).
		Env([]string{"GOROOT=" + goroot}).
		Run(origGo, "build", "-o", tmpBuiltOutput, arg)
	if err != nil {
		return err
	}

	err = os.Rename(origFile, origFile+".bak")
	if err != nil {
		return err
	}
	err = os.Rename(tmpBuiltOutput, origFile)
	if err != nil {
		return err
	}
	return nil
}
