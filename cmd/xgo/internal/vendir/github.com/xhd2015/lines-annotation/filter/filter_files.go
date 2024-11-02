package filter

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/filter"
)

func FilterFiles(project *model.ProjectAnnotation, opts *filter.Options) {
	checker := NewChecker(opts)
	if checker == nil {
		return
	}
	FilterFilesWithCheck(project, func(file model.RelativeFile) bool {
		return checker.MatchFile(string(file))
	})
}

func FilterFilesWithCheck(project *model.ProjectAnnotation, check func(file model.RelativeFile) bool) {
	if check == nil {
		return
	}
	files := project.Files
	newFiles := make(model.FileAnnotationMapping, len(files))
	for file, val := range files {
		if !check(file) {
			continue
		}
		newFiles[file] = val
	}
	project.Files = newFiles
}
