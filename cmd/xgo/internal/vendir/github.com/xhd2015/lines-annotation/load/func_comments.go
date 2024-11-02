package load

import (
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/ast"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	"github.com/xhd2015/xgo/support/fileutil"
)

func FuncCommentsFromAstInfoMapping(astLoadInfo ast.LoadInfo, mapping map[model.RelativeFile][]*model.FuncAnnotation) *model.ProjectAnnotation {
	files := make(model.FileAnnotationMapping)

	astLoadInfo.RangeFiles(func(f ast.File) bool {
		relFile := model.RelativeFile(fileutil.Slashlize(f.RelPath()))
		funcs := mapping[relFile]
		if len(funcs) == 0 {
			return true
		}
		funcAnns := make(model.FuncAnnotationMapping, len(funcs))

		lines := strings.Split(string(f.Content()), "\n")
		for _, fn := range funcs {
			var fnID model.BlockID

			lineNum := fn.Block.StartLine

			// start from previous line, checking // blocks
			// NOTE: lineNum is 1-index
			for i := lineNum - 2; i >= 0; i-- {
				line := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(line, "//") {
					break
				}
				prop, author, values := parseLineComments(line[len("//"):])
				if prop == "" {
					continue
				}
				if fnID == "" {
					fnID = fn.Block.ID()
				}
				fnAnn := funcAnns[fnID]
				if fnAnn == nil {
					fnAnn = &model.FuncAnnotation{
						Code: &model.CodeAnnotation{
							Comments: make(map[model.CommentLabel]*model.CodeComment),
						},
					}
					funcAnns[fnID] = fnAnn
				}
				cmt := fnAnn.Code.Comments[prop]
				if cmt == nil {
					cmt = &model.CodeComment{
						Author: author,
					}
					fnAnn.Code.Comments[prop] = cmt
				}
				if cmt.Author == "" {
					cmt.Author = author
				}
				for _, v := range values {
					if !strListContains(cmt.Values, v) {
						cmt.Values = append(cmt.Values, v)
					}
				}
			}
		}
		if len(funcAnns) > 0 {
			files[relFile] = &model.FileAnnotation{
				Funcs: funcAnns,
			}
		}

		return true
	})

	return &model.ProjectAnnotation{
		Files: files,
		Types: map[model.AnnotationType]bool{
			model.AnnotationType_FuncCodeComments: true,
		},
	}
}

var labelAllowValues = map[model.CommentLabel]bool{
	model.CommentLabels_Labels:      true,
	model.CommentLabels_Unreachable: false,
	model.CommentLabels_NoCov:       false,
	model.CommentLabels_Deprecated:  false,
}

func sliceStr(s string, start int, end int) string {
	if start < 0 {
		start = 0
	}
	if end < 0 {
		return s[start:]
	}
	return s[start:end]
}

// examples:
//
//	//A                              --> not parsed,prop=<empty>
//	// unreachable:reason            --> prop=unreacchable
//	// unreachable(stephen): reason  --> prop=unreacchable
//	// labels(stephen): a,b,c
//
// NOTE: only some props have values
// `line`'s prefix '//' should be trimmed
func parseLineComments(line string) (prop model.CommentLabel, author string, values []string) {
	if len(line) == 0 {
		return
	}
	if line[0] != ' ' && line[0] != '\t' {
		// expect at least one space
		return
	}
	line = strings.TrimSpace(line)

	parIdx := strings.Index(line, "(")
	colonIdx := strings.Index(line, ":")
	spaceIdx := strings.Index(line, " ")

	// has author
	if parIdx >= 0 && (colonIdx < 0 || parIdx < colonIdx) {
		subLine := sliceStr(line, parIdx+1, colonIdx)
		parEndIdx := strings.Index(subLine, ")")
		if parEndIdx >= 0 {
			author = strings.TrimSpace(subLine[:parEndIdx])
		}
	}

	propEndIdx := -1
	if parIdx >= 0 {
		propEndIdx = parIdx
	}
	if colonIdx >= 0 && (propEndIdx < 0 || propEndIdx > colonIdx) {
		propEndIdx = colonIdx
	}
	if spaceIdx >= 0 && (propEndIdx < 0 || propEndIdx > spaceIdx) {
		propEndIdx = spaceIdx
	}

	if propEndIdx == -1 {
		propEndIdx = len(line)
	}
	prop = model.CommentLabel(strings.ToLower(strings.TrimSpace(sliceStr(line, 0, propEndIdx))))
	if prop == "" {
		return
	}
	allowValues, ok := labelAllowValues[prop]
	if !ok {
		prop = ""
		return
	}

	// always covert to lowercase
	if allowValues {
		// parse values
		if colonIdx >= 0 {
			valuesRaw := strings.Split(line[colonIdx+1:], ",")
			values = make([]string, 0, len(valuesRaw))
			for _, val := range valuesRaw {
				val = strings.TrimSpace(val)
				if val != "" {
					values = append(values, val)
				}
			}
		}
	}
	return
}

func strListContains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
