package patch

import "path/filepath"

type FilePath []string

func (c FilePath) JoinPrefix(s ...string) string {
	return filepath.Join(filepath.Join(s...), filepath.Join(c...))
}

func (c FilePath) Append(s ...string) FilePath {
	return append(c, s...)
}
