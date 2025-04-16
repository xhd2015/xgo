package instrument_go

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/instrument/build"
	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/edit/goedit"
	"github.com/xhd2015/xgo/support/goinfo"
)

// src/cmd/cover/cover.go
var coverDirPath = patch.FilePath{"src", "cmd", "cover"}
var coverFilePath = coverDirPath.Append("cover.go")
var xgoCoverFilePath = coverDirPath.Append("xgo_cover.go")

//go:embed xgo_cover_template.go
var xgoCoverTemplate string

func InstrumentGoToolCover(goroot string, goVersion *goinfo.GoVersion) error {
	err := copyXgoCover(goroot)
	if err != nil {
		return err
	}
	err = instrumentCmdCover(goroot, goVersion)
	if err != nil {
		return err
	}

	toolPath, err := build.GetToolPath(goroot)
	if err != nil {
		return err
	}

	// build cover command
	return build.BuildBinary(goroot, filepath.Join(goroot, "src"), toolPath, "cover", "./cmd/cover")
}

func copyXgoCover(goroot string) error {
	code, err := patch.RemoveBuildIgnore(xgoCoverTemplate)
	if err != nil {
		return err
	}
	coverFile := xgoCoverFilePath.JoinPrefix(goroot)
	return os.WriteFile(coverFile, []byte(code), 0644)
}

func instrumentCmdCover(goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major != 1 || (goVersion.Minor < 17 || goVersion.Minor > 24) {
		// src/cmd/cover/cover.go
		return fmt.Errorf("%s unsupported version: go%d.%d, available: go1.17~go1.24", coverFilePath.JoinPrefix(""), goVersion.Major, goVersion.Minor)
	}
	coverFile := coverFilePath.JoinPrefix(goroot)

	return patch.EditFile(coverFile, func(content string) (string, error) {
		fnAnchor := "\nfunc (p *Package) annotateFile(name string," // since go1.20
		if goVersion.Minor == 17 || goVersion.Minor == 18 || goVersion.Minor == 19 {
			fnAnchor = "\nfunc annotate(name string"
		}
		content = patch.UpdateContent(content,
			"/*<begin get_xgo_edits>*/",
			"/*<end get_xgo_edits>*/",
			[]string{
				fnAnchor,
				"parsedFile, err := parser.ParseFile(",
			},
			1,
			patch.UpdatePosition_Before,
			"__xgo_edits := xgoParseApply(&name, &content);",
		)
		content = patch.UpdateContent(content,
			"/*<begin apply_xgo_edits>*/",
			"/*<end apply_xgo_edits>*/",
			[]string{
				fnAnchor,
				"parsedFile, err := parser.ParseFile(",
				"newContent := file.edit.Bytes()",
			},
			2,
			patch.UpdatePosition_Before,
			"for _, e := range __xgo_edits { file.edit.Replace(e.Start, e.End, e.New); };",
		)
		return content, nil
	})

}

// AddEditsNotes adds notes to the content to indicate the edits made to the file.
// It marshals the edits into JSON and appends it to the content.
// to support the solution described in https://github.com/xhd2015/xgo/issues/301
func AddEditsNotes(fileEdit *goedit.Edit, file string, origalContent string, content string) (string, error) {
	type Edit struct {
		Start int    `json:"start,omitempty"`
		End   int    `json:"end,omitempty"`
		New   string `json:"new,omitempty"`
	}
	var edits []Edit
	fileEdit.Buffer().RangeEdits(func(start int, end int, new string) {
		edits = append(edits, Edit{Start: start, End: end, New: new})
	})
	editsJSON, err := json.Marshal(edits)
	if err != nil {
		return "", err
	}
	sep := ""
	if !strings.HasSuffix(content, "\n") {
		sep = "\n"
	}
	return content + sep +
		"// __xgo_file: " + strconv.Quote(file) + "\n" +
		"// __xgo_edits: " + string(editsJSON) + "\n" +
		"// __xgo_original: " + strconv.Quote(origalContent), nil
}
