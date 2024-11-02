package git

import "strings"

//  git branch --list -r
//
//	origin/dev-1.0.0
//	origin/master
func ListBranch(dir string) ([]string, error) {
	res, err := RunCommand(dir, func(commands []string) []string {
		return append(commands, []string{
			"git branch --list -r --sort=-committerdate",
		}...)
	})
	if err != nil {
		return nil, err
	}

	lines := splitLinesFilterEmpty(res)
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "origin/") {
			continue
		}
		// ref branch
		if strings.Contains(line, "->") {
			continue
		}
		branch := strings.TrimPrefix(line, "origin/")
		branches = append(branches, branch)
	}

	return branches, nil
}
