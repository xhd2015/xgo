package instrument_go

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var execFilePath = internalWorkPath.Append("exec.go")               // src/cmd/go/internal/work/exec.go
var xgoWorkSumFilePath = internalWorkPath.Append("xgo_work_sum.go") // src/cmd/go/internal/work/xgo_work_sum.go

// instrumentExec instrument the internal exec.go to fix coverage with -overlay
func instrumentExec(goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 25) {
		// src/cmd/go/internal/work/exec.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.25", execFilePath.JoinPrefix(""), goVersion.Major, goVersion.Minor)
	}
	execFile := execFilePath.JoinPrefix(goroot)

	return patch.EditFile(execFile, func(content string) (string, error) {
		// since 22
		coverLine := `if p.Internal.Cover.Mode != "" {`
		getActualFile := `__xgo_overlay_source_file := sourceFile; if _actual := fsys.Actual(sourceFile); _actual != "" { __xgo_overlay_source_file = _actual; }; ` // since 24
		if goVersion.Minor < 24 {
			getActualFile = "__xgo_overlay_source_file, _ := fsys.OverlayPath(sourceFile);"
		}
		// In go1.25, b.cover call changed from "if err := b.cover(" to "if newoutfiles, err := b.cover("
		coverCallAnchor := "if err := b.cover("
		if goVersion.Minor >= 25 {
			coverCallAnchor = "if newoutfiles, err := b.cover("
		}
		switch goVersion.Minor {
		case 20, 21:
			coverLine = `if p.Internal.CoverMode != "" {`
		case 18, 19, 17:
			coverLine = `if a.Package.Internal.CoverMode != "" {`
		}
		// In go1.25, sourceFile is no longer passed directly to b.cover - it goes
		// through infiles. So the var_declare and var_replace patches only apply to <go1.25.
		if goVersion.Minor < 25 {
			content = patch.UpdateContent(content,
				"/*<begin fix_cover_overlay_var_declare>*/",
				"/*<end fix_cover_overlay_var_declare>*/",
				[]string{
					"func (b *Builder) build(",
					coverLine,
					coverCallAnchor,
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
					coverCallAnchor,
					"sourceFile,",
				},
				3,
				patch.UpdatePosition_Replace,
				"__xgo_overlay_source_file,",
			)
		}
		// redesign
		if goVersion.Minor >= 20 && goVersion.Minor < 25 {
			// In go1.25, CoverageRedesign is the default and the cfg.Experiment check was removed.
			// The infiles append is now unconditional within the cover mode block.
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
		if goVersion.Minor >= 25 {
			// In go1.25, CoverageRedesign is always on, so infiles append is unconditional
			content = patch.UpdateContent(content,
				"/*<begin modify_infiles>*/",
				"/*<end modify_infiles>*/",
				[]string{
					"func (b *Builder) build(",
					coverLine,
					"infiles = append(infiles, sourceFile)",
				},
				2,
				patch.UpdatePosition_Replace,
				strings.TrimSuffix(getActualFile, ";")+";"+"infiles = append(infiles, __xgo_overlay_source_file)",
			)
		}
		if true {
			content = execAddXgoTrapSum(content)
		}
		return content, nil
	})
}

// Deprecated: this affects build cache
// slows down build
func execAddXgoTrapSum(content string) string {
	// only for debugging
	content = patch.UpdateContent(content,
		"/*<begin add_xgo_trap_sum>*/",
		"/*<end add_xgo_trap_sum>*/",
		[]string{
			"func (b *Builder) buildActionID",
			`fmt.Fprintf(h, "compile\n")`,
		},
		1,
		patch.UpdatePosition_After,
		`;if xgoSum := getXgoPackageTrapSum(p.ImportPath); xgoSum!="" { fmt.Fprintf(h, "xgo trap sum %s\n", xgoSum);}`,
	)
	return content
}
