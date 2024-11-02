package compute

import (
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
)

func LineLabelsFromLineComment(p *model.ProjectAnnotation) {
	for file, fileAn := range p.Files {
		fileData := &model.FileAnnotation{
			Lines: make(model.LineAnnotationMapping),
		}
		for line, lineAn := range fileAn.Lines {
			if lineAn.Code == nil {
				continue
			}
			cmt := lineAn.Code.Comments[model.CommentLabels_Labels]
			if cmt == nil {
				continue
			}
			var labels map[string]bool
			for _, val := range cmt.Values {
				if labels == nil {
					labels = make(map[string]bool)
				}
				labels[strings.ToUpper(val)] = true
			}
			if len(labels) == 0 {
				continue
			}
			fileData.Lines[line] = &model.LineAnnotation{
				Labels: labels,
			}
		}
		p.Files[file] = fileData
	}
}
