package git

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/git"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/line"
	cov_model "github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-coverage/model"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model"
	"github.com/xhd2015/xgo/support/fileutil"
)

func LoadFiles(dir string, ref string, diffBase string, relFiles []string) (*model.ProjectAnnotation, error) {
	gitFiles, err := git.DiffCommitFiles(dir, ref, diffBase, nil)
	if err != nil {
		return nil, err
	}

	files := make(model.FileAnnotationMapping, len(relFiles))
	for _, relFile := range relFiles {
		gitFile := gitFiles[relFile]
		if gitFile == nil {
			// unknown file
			continue
		}
		if gitFile.Deleted {
			continue
		}

		var lineChanges *cov_model.LineChanges
		if !gitFile.IsNew && gitFile.ContentChanged {
			oldFile := relFile
			if gitFile.RenamedFrom != "" {
				oldFile = gitFile.RenamedFrom
			}
			var err error
			lineChanges, err = diffFile(dir, ref, relFile, diffBase, oldFile)
			if err != nil {
				return nil, err
			}
		}

		files[model.RelativeFile(fileutil.Slashlize(relFile))] = &model.FileAnnotation{
			ChangeDetail: &cov_model.FileDetail{
				IsNew:          gitFile.IsNew,
				Deleted:        gitFile.Deleted,
				RenamedFrom:    gitFile.RenamedFrom,
				ContentChanged: gitFile.ContentChanged,
			},
			LineChanges: lineChanges,
		}
	}

	return &model.ProjectAnnotation{
		Files: files,
		Types: map[model.AnnotationType]bool{
			model.AnnotationType_LineChanges:  true,
			model.AnnotationType_ChangeDetail: true,
		},
	}, nil
}

func diffFile(dir string, ref string, relFile string, diffBase string, oldFile string) (*cov_model.LineChanges, error) {
	ok, newContent, err := git.CatFile(dir, ref, relFile)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unknown file: %s", relFile)
	}
	if oldFile == "" {
		oldFile = relFile
	}
	newContentStr := string(newContent)
	ok, oldContent, err := git.CatFile(dir, diffBase, oldFile)
	if err != nil {
		return nil, err
	}
	var lineChanges *cov_model.LineChanges
	if ok && !isContentTrimSpaceSame(newContentStr, oldContent) {
		lineChanges, err = line.Diff(newContentStr, oldContent)
		if err != nil {
			return nil, fmt.Errorf("diff %s: %w", relFile, err)
		}
	}

	return lineChanges, nil
}

func isContentTrimSpaceSame(a string, b string) bool {
	return strings.TrimSpace(a) == strings.TrimSpace(b)
}
