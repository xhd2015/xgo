package filter

import (
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/filter"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/lang"
)

func MakeFilter(profileLang lang.ProfileLanguage) *filter.Options {
	if profileLang.IsGo() {
		return &filter.Options{
			Suffix:        []string{".go"},
			ExcludeSuffix: []string{"_test.go"},
			Exclude:       []string{"vendor"},
		}
	}
	if profileLang == lang.ProfileLanguage_Js {
		return &filter.Options{
			Suffix:  []string{".js", ".ts"},
			Exclude: []string{"node_modules"},
		}
	}
	return nil
}
