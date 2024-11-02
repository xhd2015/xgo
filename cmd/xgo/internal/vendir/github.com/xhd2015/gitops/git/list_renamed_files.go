package git

type FileRename struct {
	File       string
	RenameFrom string
	Percent    string // e.g. 100%
}

func ListRenamedFiles(dir string, ref string, compareRef string, patterns []string) ([]*FileRename, error) {
	content, err := diffFiles(dir, ref, compareRef, []string{"--diff-filter=R", "--summary"}, patterns)
	if err != nil {
		return nil, err
	}
	var renames []*FileRename
	// check renames and updates
	// iterate for each 'rename ' at line begin
	// do not consider filename containing newline,"{" and "}", and space in the end or start
	parseRenames(content, func(newFile, oldFile, percent string) {
		renames = append(renames, &FileRename{
			File:       newFile,
			RenameFrom: oldFile,
			Percent:    percent,
		})
	})

	return renames, nil
}
