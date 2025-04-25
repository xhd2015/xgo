package patch

import "path/filepath"

type FilePath []string

func (c FilePath) JoinPrefix(s ...string) string {
	return filepath.Join(filepath.Join(s...), filepath.Join(c...))
}

func (c FilePath) Append(s ...string) FilePath {
	clone := make(FilePath, len(c)+len(s))
	copy(clone, c)
	copy(clone[len(c):], s)
	return clone
}
