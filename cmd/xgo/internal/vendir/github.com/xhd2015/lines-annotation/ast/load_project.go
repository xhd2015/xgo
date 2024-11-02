package ast

import "github.com/xhd2015/xgo/support/goinfo"

func LoadProject(dir string, args []string) (LoadInfo, error) {
	files, err := goinfo.ListRelativeFiles(dir, args)
	if err != nil {
		return nil, err
	}
	return LoadFiles(dir, files)
}
