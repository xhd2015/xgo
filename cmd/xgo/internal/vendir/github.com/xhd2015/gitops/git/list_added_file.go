package git

func ListAddedFile(dir string, ref string, compareRef string, patterns []string) ([]string, error) {
	content, err := diffFiles(dir, ref, compareRef, []string{"--diff-filter=A"}, patterns)
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(content), nil
}
