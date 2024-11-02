package git

func ListModifiedFiles(dir string, ref string, compareRef string, patterns []string) ([]string, error) {
	content, err := diffFiles(dir, ref, compareRef, []string{"--diff-filter=M"}, patterns)
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(content), nil
}
